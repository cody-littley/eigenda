package relay

import (
	"context"
	"github.com/Layr-Labs/eigenda/common"
	tu "github.com/Layr-Labs/eigenda/common/testutils"
	v2 "github.com/Layr-Labs/eigenda/core/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadWrite(t *testing.T) {
	tu.InitializeRandom()

	logger, err := common.NewLogger(common.DefaultLoggerConfig())
	require.NoError(t, err)

	setup(t)

	blobStore := buildBlobStore(t, logger)

	expectedData := make(map[v2.BlobKey][]byte)

	blobCount := 10
	for i := 0; i < blobCount; i++ {
		header, data := randomBlob(t)

		blobKey, err := header.BlobKey()
		require.NoError(t, err)
		expectedData[blobKey] = data

		err = blobStore.StoreBlob(context.Background(), blobKey, data)
		require.NoError(t, err)
	}

	server, err := NewBlobServer(context.Background(), logger, blobStore, 10, 32)
	require.NoError(t, err)

	// Read the blobs back.
	for key, data := range expectedData {
		blob, err := server.GetBlob(key)

		require.NoError(t, err)
		require.Equal(t, data, blob)
	}

	// Read the blobs back again to test caching.
	for key, data := range expectedData {
		blob, err := server.GetBlob(key)

		require.NoError(t, err)
		require.Equal(t, data, blob)
	}
}
