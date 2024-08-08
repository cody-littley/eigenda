package workers

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/Layr-Labs/eigenda/common"
	tu "github.com/Layr-Labs/eigenda/common/testutils"
	"github.com/Layr-Labs/eigenda/encoding/utils/codec"
	"github.com/Layr-Labs/eigenda/tools/traffic/config"
	"github.com/Layr-Labs/eigenda/tools/traffic/metrics"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
	"sync"
	"testing"
)

func TestBlobWriter(t *testing.T) {
	tu.InitializeRandom()

	ctx, cancel := context.WithCancel(context.Background())
	waitGroup := sync.WaitGroup{}
	logger, err := common.NewLogger(common.DefaultLoggerConfig())
	assert.Nil(t, err)

	dataSize := rand.Uint64()%1024 + 64

	authenticated := rand.Intn(2) == 0
	var signerPrivateKey string
	if authenticated {
		signerPrivateKey = "asdf"
	}

	randomizeBlobs := rand.Intn(2) == 0

	useCustomQuorum := rand.Intn(2) == 0
	var customQuorum []uint8
	if useCustomQuorum {
		customQuorum = []uint8{1, 2, 3}
	}

	config := &config.WorkerConfig{
		DataSize:         dataSize,
		SignerPrivateKey: signerPrivateKey,
		RandomizeBlobs:   randomizeBlobs,
		CustomQuorums:    customQuorum,
	}

	disperserClient := NewMockDisperserClient(t, authenticated)
	unconfirmedKeyHandler := NewMockKeyHandler(t)

	generatorMetrics := metrics.NewMockMetrics()

	writer := NewBlobWriter(
		&ctx,
		&waitGroup,
		logger,
		config,
		disperserClient,
		unconfirmedKeyHandler,
		generatorMetrics)

	errorCount := 0

	var previousData []byte

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			disperserClient.DispenseErrorToReturn = fmt.Errorf("intentional error for testing purposes")
			errorCount++
		} else {
			disperserClient.DispenseErrorToReturn = nil
		}

		// This is the key that will be assigned to the next blob.
		disperserClient.KeyToReturn = make([]byte, 32)
		_, err = rand.Read(disperserClient.KeyToReturn)
		assert.Nil(t, err)

		// Simulate the advancement of time (i.e. allow the writer to write the next blob).
		writer.writeNextBlob()

		assert.Equal(t, uint(i+1), disperserClient.DisperseCount)
		assert.Equal(t, uint(i+1-errorCount), unconfirmedKeyHandler.Count)

		// This method should not be called in this test.
		assert.Equal(t, uint(0), disperserClient.GetStatusCount)

		if disperserClient.DispenseErrorToReturn == nil {
			assert.NotNil(t, disperserClient.ProvidedData)
			assert.Equal(t, customQuorum, disperserClient.ProvidedQuorum)

			// Strip away the extra encoding bytes. We should have data of the expected size.
			decodedData := codec.RemoveEmptyByteFromPaddedBytes(disperserClient.ProvidedData)
			assert.Equal(t, dataSize, uint64(len(decodedData)))

			// Verify that the proper data was sent to the unconfirmed key handler.
			assert.Equal(t, uint(len(disperserClient.ProvidedData)), unconfirmedKeyHandler.ProvidedSize)
			checksum := md5.Sum(disperserClient.ProvidedData)

			assert.Equal(t, checksum, unconfirmedKeyHandler.ProvidedChecksum)
			assert.Equal(t, disperserClient.KeyToReturn, unconfirmedKeyHandler.ProvidedKey)

			//assert.Equal(t, checksum, unconfirmedKeyHandler.ProvidedChecksum)
			//assert.Equal(t, disperserClient.KeyToReturn, unconfirmedKeyHandler.ProvidedKey)

			// Verify that data has the proper amount of randomness.
			if previousData != nil {
				if randomizeBlobs {
					// We expect each blob to be different.
					assert.NotEqual(t, previousData, disperserClient.ProvidedData)
				} else {
					// We expect each blob to be the same.
					assert.Equal(t, previousData, disperserClient.ProvidedData)
				}
			}
			previousData = disperserClient.ProvidedData
		}

		// Verify metrics.
		assert.Equal(t, float64(i+1-errorCount), generatorMetrics.GetCount("write_success"))
		assert.Equal(t, float64(errorCount), generatorMetrics.GetCount("write_failure"))
	}

	cancel()
}
