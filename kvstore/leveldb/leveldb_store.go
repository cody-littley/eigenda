package leveldb

import (
	"errors"
	"fmt"
	"github.com/Layr-Labs/eigenda/kvstore"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var _ kvstore.Store = &levelDBStore{}

// levelDBStore implements kvstore.Store interfaces with levelDB as the backend engine.
type levelDBStore struct {
	db   *leveldb.DB
	path string

	logger logging.Logger

	shutdown  bool
	destroyed bool
}

// NewStore returns a new levelDBStore built using LevelDB.
func NewStore(logger logging.Logger, path string) (kvstore.Store, error) {
	levelDB, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return nil, err
	}

	return &levelDBStore{
		db:     levelDB,
		logger: logger,
	}, nil
}

func (store *levelDBStore) Put(key []byte, value []byte) error {
	return store.db.Put(key, value, nil)
}

func (store *levelDBStore) Get(key []byte) ([]byte, error) {
	data, err := store.db.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, kvstore.ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

func (store *levelDBStore) NewIterator(prefix []byte) iterator.Iterator {
	return store.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (store *levelDBStore) Delete(key []byte) error {
	err := store.db.Delete(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return kvstore.ErrNotFound
		}
		return nil
	}
	return nil
}

func (store *levelDBStore) DeleteBatch(keys [][]byte) error {
	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete(key)
	}
	return store.db.Write(batch, nil)
}

func (store *levelDBStore) WriteBatch(keys, values [][]byte) error {
	batch := new(leveldb.Batch)
	for i, key := range keys {
		batch.Put(key, values[i])
	}
	return store.db.Write(batch, nil)
}

// Shutdown shuts down the store.
func (store *levelDBStore) Shutdown() error {
	if store.shutdown {
		return nil
	}

	err := store.db.Close()
	if err != nil {
		return err
	}

	store.shutdown = true
	return nil
}

// Destroy destroys the store.
func (store *levelDBStore) Destroy() error {
	if store.destroyed {
		return nil
	}

	if !store.shutdown {
		err := store.Shutdown()
		if err != nil {
			return err
		}
	}

	store.logger.Info(fmt.Sprintf("destroying LevelDB store at path: %s", store.path))
	err := os.RemoveAll(store.path)
	if err != nil {
		return err
	}
	store.destroyed = true
	return nil
}
