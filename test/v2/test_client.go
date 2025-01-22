package v2

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/clients/v2"
	commonv2 "github.com/Layr-Labs/eigenda/api/grpc/common/v2"
	v2 "github.com/Layr-Labs/eigenda/api/grpc/disperser/v2"
	"github.com/Layr-Labs/eigenda/common"
	"github.com/Layr-Labs/eigenda/common/geth"
	"github.com/Layr-Labs/eigenda/core"
	auth "github.com/Layr-Labs/eigenda/core/auth/v2"
	"github.com/Layr-Labs/eigenda/core/eth"
	"github.com/Layr-Labs/eigenda/core/thegraph"
	corev2 "github.com/Layr-Labs/eigenda/core/v2"
	"github.com/Layr-Labs/eigenda/encoding/kzg"
	"github.com/Layr-Labs/eigenda/encoding/kzg/verifier"
	"github.com/Layr-Labs/eigenda/encoding/utils/codec"
	"github.com/Layr-Labs/eigensdk-go/logging"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

// TestClient encapsulates the various clients necessary for interacting with EigenDA.
type TestClient struct {
	t                 *testing.T
	logger            logging.Logger
	DisperserClient   clients.DisperserClient
	RelayClient       clients.RelayClient
	indexedChainState core.IndexedChainState
	RetrievalClient   clients.RetrievalClient
}

type TestClientConfig struct {
	PrivateKeyFile                string
	DisperserHostname             string
	DisperserPort                 int
	EthRPCURLs                    []string
	BLSOperatorStateRetrieverAddr string
	EigenDAServiceManagerAddr     string
	SubgraphURL                   string
	KZGPath                       string
	SRSOrder                      uint64
	SRSNumberToLoad               uint64
}

// NewTestClient creates a new TestClient instance.
func NewTestClient(t *testing.T, config *TestClientConfig) *TestClient {

	logger, err := common.NewLogger(common.DefaultTextLoggerConfig())
	require.NoError(t, err)

	// Construct the disperser client

	privateKeyFile := config.PrivateKeyFile
	privateKey, err := os.ReadFile(privateKeyFile)
	require.NoError(t, err)

	privateKeyString := string(privateKey)
	privateKeyString = strings.Trim(privateKeyString, "\n \t")
	privateKeyString, _ = strings.CutPrefix(privateKeyString, "0x")

	signer := auth.NewLocalBlobRequestSigner(privateKeyString)
	signerAccountId, err := signer.GetAccountID()
	require.NoError(t, err)
	accountId := gethcommon.HexToAddress(signerAccountId)
	fmt.Printf("Account ID: %s\n", accountId.String())

	disperserConfig := &clients.DisperserClientConfig{
		Hostname:          config.DisperserHostname,
		Port:              fmt.Sprintf("%d", config.DisperserPort),
		UseSecureGrpcFlag: true,
	}
	disperserClient, err := clients.NewDisperserClient(disperserConfig, signer, nil, nil)
	require.NoError(t, err)

	// Construct the relay client

	ethClientConfig := geth.EthClientConfig{
		RPCURLs:          config.EthRPCURLs,
		PrivateKeyString: privateKeyString,
		NumConfirmations: 0,
		NumRetries:       3,
	}
	ethClient, err := geth.NewMultiHomingClient(ethClientConfig, accountId, logger)
	require.NoError(t, err)

	ethReader, err := eth.NewReader(
		logger,
		ethClient,
		config.BLSOperatorStateRetrieverAddr,
		config.EigenDAServiceManagerAddr)
	require.NoError(t, err)

	relayURLS, err := ethReader.GetRelayURLs(context.Background())
	require.NoError(t, err)

	relayConfig := &clients.RelayClientConfig{
		Sockets:           relayURLS,
		UseSecureGrpcFlag: true,
	}
	relayClient, err := clients.NewRelayClient(relayConfig, logger)
	require.NoError(t, err)

	// Construct the retrieval client

	chainState := eth.NewChainState(ethReader, ethClient)
	icsConfig := thegraph.Config{
		Endpoint:     config.SubgraphURL,
		PullInterval: 100 * time.Millisecond,
		MaxRetries:   5,
	}
	indexedChainState := thegraph.MakeIndexedChainState(icsConfig, chainState, logger)

	kzgConfig := &kzg.KzgConfig{
		LoadG2Points:    true,
		G1Path:          path.Join(config.KZGPath, "g1.point"),
		G2Path:          path.Join(config.KZGPath, "g2.point"),
		G2PowerOf2Path:  path.Join(config.KZGPath, "g2.point.powerOf2"),
		CacheDir:        path.Join(config.KZGPath, "SRSTables"),
		SRSOrder:        config.SRSOrder,
		SRSNumberToLoad: config.SRSNumberToLoad,
		NumWorker:       32,
	}
	blobVerifier, err := verifier.NewVerifier(kzgConfig, nil)
	require.NoError(t, err)

	retrievalClient := clients.NewRetrievalClient(
		logger,
		ethReader,
		indexedChainState,
		blobVerifier,
		20)

	return &TestClient{
		t:                 t,
		logger:            logger,
		DisperserClient:   disperserClient,
		RelayClient:       relayClient,
		indexedChainState: indexedChainState,
		RetrievalClient:   retrievalClient,
	}
}

// DisperseAndVerify sends a payload to the disperser. Waits until the payload is confirmed and then reads
// it back from the relays and the validators.
func (c *TestClient) DisperseAndVerify(
	ctx context.Context,
	payload []byte,
	quorums []core.QuorumID) {

	key := c.DispersePayload(ctx, payload, quorums)
	blobCert := c.WaitForCertification(ctx, key)
	c.ReadBlobFromRelay(ctx, key, blobCert, payload)
	c.ReadBlobFromValidators(ctx, blobCert, quorums, payload)
}

// DispersePayload sends a payload to the disperser. Returns the blob key.
func (c *TestClient) DispersePayload(
	ctx context.Context,
	payload []byte,
	quorums []core.QuorumID) corev2.BlobKey {

	fmt.Printf("Dispersing payload of length %d to quorums %v\n", len(payload), quorums)

	padded := codec.ConvertByPaddingEmptyByte(payload)
	_, key, err := c.DisperserClient.DisperseBlob(ctx, padded, 0, quorums, 0)
	require.NoError(c.t, err)
	fmt.Printf("Dispersed blob with key %x\n", key)

	return key
}

// WaitForCertification waits for a blob to be certified. Returns the blob certificate.
func (c *TestClient) WaitForCertification(ctx context.Context, key corev2.BlobKey) *commonv2.BlobCertificate {
	var status *v2.BlobStatus = nil
	ticker := time.NewTicker(time.Second)
	start := time.Now()
	statusStart := start
	for {
		select {
		case <-ticker.C:
			reply, err := c.DisperserClient.GetBlobStatus(ctx, key)
			require.NoError(c.t, err)

			if reply.Status == v2.BlobStatus_CERTIFIED {
				elapsed := time.Since(statusStart)
				totalElapsed := time.Since(start)
				fmt.Printf(
					"Blob is certified (spent %0.1fs in prior status, total time %0.1fs)\n",
					elapsed.Seconds(),
					totalElapsed.Seconds())

				blobCert := reply.BlobVerificationInfo.BlobCertificate
				require.NotNil(c.t, blobCert)
				require.True(c.t, len(blobCert.Relays) >= 1)
				return blobCert

			} else if status == nil || reply.Status != *status {
				elapsed := time.Since(statusStart)
				statusStart = time.Now()
				if status == nil {
					fmt.Printf("Blob status: %s\n", reply.Status.String())
				} else {
					fmt.Printf("Blob status: %s (spent %0.1fs in prior status)\n",
						reply.Status.String(),
						elapsed.Seconds())
				}
				status = &reply.Status

				if reply.Status == v2.BlobStatus_FAILED ||
					reply.Status == v2.BlobStatus_UNKNOWN ||
					reply.Status == v2.BlobStatus_INSUFFICIENT_SIGNATURES {
					require.Fail(
						c.t,
						"Blob status is in a terminal non-successful state.",
						reply.Status.String())
				}
			}
		case <-ctx.Done():
			require.Fail(c.t, "Timed out waiting for blob to be confirmed")
		}
	}
}

// ReadBlobFromRelay reads a blob from the relays and compares it to the given payload.
func (c *TestClient) ReadBlobFromRelay(
	ctx context.Context,
	key corev2.BlobKey,
	blobCert *commonv2.BlobCertificate,
	payload []byte) {

	for _, relayID := range blobCert.Relays {
		fmt.Printf("Reading blob from relay %d\n", relayID)
		blobFromRelay, err := c.RelayClient.GetBlob(ctx, relayID, key)
		require.NoError(c.t, err)

		relayPayload := codec.RemoveEmptyByteFromPaddedBytes(blobFromRelay)
		require.Equal(c.t, payload, relayPayload)
	}
}

// ReadBlobFromValidators reads a blob from the validators and compares it to the given payload.
func (c *TestClient) ReadBlobFromValidators(
	ctx context.Context,
	blobCert *commonv2.BlobCertificate,
	quorums []core.QuorumID,
	payload []byte) {

	currentBlockNumber, err := c.indexedChainState.GetCurrentBlockNumber()
	require.NoError(c.t, err)

	for _, quorumID := range quorums {
		fmt.Printf("Reading blob from validators for quorum %d\n", quorumID)
		header, err := corev2.BlobHeaderFromProtobuf(blobCert.BlobHeader)
		require.NoError(c.t, err)

		retrievedBlob, err := c.RetrievalClient.GetBlob(ctx, header, uint64(currentBlockNumber), quorumID)
		require.NoError(c.t, err)

		retrievedPayload := codec.RemoveEmptyByteFromPaddedBytes(retrievedBlob)

		// The payload may have a bunch of 0s appended at the end. Remove them.
		require.True(c.t, len(retrievedPayload) >= len(payload))
		truncatedPayload := retrievedPayload[:len(payload)]

		// Only 0s should be appended at the end.
		for i := len(payload); i < len(retrievedPayload); i++ {
			require.Equal(c.t, byte(0), retrievedPayload[i])
		}

		require.Equal(c.t, payload, truncatedPayload)
	}
}
