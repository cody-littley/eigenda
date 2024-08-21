package test

import (
	"github.com/Layr-Labs/eigenda/common"
	tu "github.com/Layr-Labs/eigenda/common/testutils"
	"github.com/Layr-Labs/eigenda/kvstore"
	"github.com/Layr-Labs/eigenda/kvstore/leveldb"
	"github.com/Layr-Labs/eigenda/kvstore/mapstore"
	"github.com/Layr-Labs/eigenda/kvstore/storeutil"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
)

// A list of builders for various stores to be tested.
var storeBuilders = []func(logger logging.Logger, path string) (kvstore.Store, error){
	func(logger logging.Logger, path string) (kvstore.Store, error) {
		return mapstore.NewStore(), nil
	},
	func(logger logging.Logger, path string) (kvstore.Store, error) {
		return storeutil.ThreadSafeWrapper(mapstore.NewStore()), nil
	},
	func(logger logging.Logger, path string) (kvstore.Store, error) {
		return leveldb.NewStore(logger, path)
	},
	func(logger logging.Logger, path string) (kvstore.Store, error) {
		store, err := leveldb.NewStore(logger, path)
		if err != nil {
			return nil, err
		}
		return storeutil.ThreadSafeWrapper(store), nil
	},
}

var dbPath = "test-store"

func deleteDBDirectory(t *testing.T) {
	err := os.RemoveAll(dbPath)
	assert.NoError(t, err)
}

func verifyDBIsDeleted(t *testing.T) {
	_, err := os.Stat(dbPath)
	assert.True(t, os.IsNotExist(err))
}

func randomOperationsTest(t *testing.T, store kvstore.Store) {
	tu.InitializeRandom()
	deleteDBDirectory(t)

	expectedData := make(map[string][]byte)

	for i := 0; i < 1000; i++ {

		choice := rand.Float64()
		if len(expectedData) == 0 || choice < 0.50 {
			// Write a random value.

			key := tu.RandomBytes(32)
			value := tu.RandomBytes(32)

			err := store.Put(key, value)
			assert.NoError(t, err)

			expectedData[string(key)] = value
		} else if choice < 0.75 {
			// Modify a random value.

			var key string
			for k := range expectedData {
				key = k
				break
			}
			value := tu.RandomBytes(32)
			err := store.Put([]byte(key), value)
			assert.NoError(t, err)
			expectedData[key] = value
		} else if choice < 0.90 {
			// Drop a random value.

			var key string
			for k := range expectedData {
				key = k
				break
			}
			delete(expectedData, key)
			err := store.Delete([]byte(key))
			assert.NoError(t, err)
		} else {
			// Drop a non-existent value.

			key := tu.RandomBytes(32)
			err := store.Delete(key)
			assert.Nil(t, err)
		}

		if i%10 == 0 {
			// Every so often, check that the store matches the expected data.
			for key, expectedValue := range expectedData {
				value, err := store.Get([]byte(key))
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, value)
			}

			// Try and get a value that isn't in the store.
			key := tu.RandomBytes(32)
			value, err := store.Get(key)
			assert.Equal(t, kvstore.ErrNotFound, err)
			assert.Nil(t, value)
		}
	}

	err := store.Shutdown()
	assert.NoError(t, err)
	err = store.Destroy()
	assert.NoError(t, err)
	verifyDBIsDeleted(t)
}

func TestRandomOperations(t *testing.T) {
	logger, err := common.NewLogger(common.DefaultLoggerConfig())
	assert.NoError(t, err)

	for _, builder := range storeBuilders {
		store, err := builder(logger, dbPath)
		assert.NoError(t, err)
		randomOperationsTest(t, store)
	}
}

func writeBatchTest(t *testing.T, store kvstore.Store) {
	tu.InitializeRandom()
	deleteDBDirectory(t)

	var err error

	expectedData := make(map[string][]byte)

	keys := make([][]byte, 0)
	values := make([][]byte, 0)

	for i := 0; i < 1000; i++ {

		// Write a random value.

		key := tu.RandomBytes(32)
		value := tu.RandomBytes(32)

		keys = append(keys, key)
		values = append(values, value)

		expectedData[string(key)] = value

		if i%10 == 0 {
			// Every so often, apply the batch and check that the store matches the expected data.

			err := store.WriteBatch(keys, values)
			assert.NoError(t, err)

			keys = make([][]byte, 0)
			values = make([][]byte, 0)

			for key, expectedValue := range expectedData {
				value, err := store.Get([]byte(key))
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, value)
			}

			// Try and get a value that isn't in the store.
			key := tu.RandomBytes(32)
			value, err := store.Get(key)
			assert.Equal(t, kvstore.ErrNotFound, err)
			assert.Nil(t, value)
		}
	}

	err = store.Shutdown()
	assert.NoError(t, err)
	err = store.Destroy()
	assert.NoError(t, err)
	verifyDBIsDeleted(t)
}

func TestWriteBatch(t *testing.T) {
	logger, err := common.NewLogger(common.DefaultLoggerConfig())
	assert.NoError(t, err)

	for _, builder := range storeBuilders {
		store, err := builder(logger, dbPath)
		assert.NoError(t, err)
		writeBatchTest(t, store)
	}
}

func deleteBatchTest(t *testing.T, store kvstore.Store) {
	tu.InitializeRandom()
	deleteDBDirectory(t)

	expectedData := make(map[string][]byte)

	keys := make([][]byte, 0)

	// Add some data to the store.
	for i := 0; i < 1000; i++ {
		key := tu.RandomBytes(32)
		value := tu.RandomBytes(32)

		err := store.Put(key, value)
		assert.NoError(t, err)

		expectedData[string(key)] = value
	}

	// Delete some of the data.
	for key := range expectedData {
		choice := rand.Float64()
		if choice < 0.5 {
			keys = append(keys, []byte(key))
			delete(expectedData, key)
		} else if choice < 0.75 {
			// Delete a non-existent key.
			keys = append(keys, tu.RandomBytes(32))
		}
	}

	err := store.DeleteBatch(keys)
	assert.NoError(t, err)

	// Check that the store matches the expected data.
	for key, expectedValue := range expectedData {
		value, err := store.Get([]byte(key))
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// Try and get a value that isn't in the store.
	key := tu.RandomBytes(32)
	value, err := store.Get(key)
	assert.Equal(t, kvstore.ErrNotFound, err)
	assert.Nil(t, value)

	err = store.Shutdown()
	assert.NoError(t, err)
	err = store.Destroy()
	assert.NoError(t, err)

	verifyDBIsDeleted(t)
}

func TestDeleteBatch(t *testing.T) {
	logger, err := common.NewLogger(common.DefaultLoggerConfig())
	assert.NoError(t, err)

	for _, builder := range storeBuilders {
		store, err := builder(logger, dbPath)
		assert.NoError(t, err)
		deleteBatchTest(t, store)
	}
}

func operationsOnShutdownStoreTest(t *testing.T, store kvstore.Store) {
	deleteDBDirectory(t)
	err := store.Shutdown()
	assert.NoError(t, err)

	err = store.Put([]byte("key"), []byte("value"))
	assert.Error(t, err)

	_, err = store.Get([]byte("key"))
	assert.Error(t, err)

	err = store.Delete([]byte("key"))
	assert.Error(t, err)

	err = store.WriteBatch(make([][]byte, 0), make([][]byte, 0))
	assert.Error(t, err)

	err = store.DeleteBatch(make([][]byte, 0))
	assert.Error(t, err)

	err = store.Shutdown()
	assert.NoError(t, err)

	err = store.Destroy()
	assert.NoError(t, err)
	verifyDBIsDeleted(t)
}

func TestOperationsOnShutdownStore(t *testing.T) {
	logger, err := common.NewLogger(common.DefaultLoggerConfig())
	assert.NoError(t, err)

	for _, builder := range storeBuilders {
		store, err := builder(logger, dbPath)
		assert.NoError(t, err)
		operationsOnShutdownStoreTest(t, store)
	}
}
