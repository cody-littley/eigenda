package tablestore

import (
	"encoding/binary"
	"github.com/Layr-Labs/eigenda/common/kvstore"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

var _ kvstore.Table = &tableView{}

// tableView allows table in a TableStore to be accessed as if it were a Store.
type tableView struct {
	// base is the underlying store.
	base kvstore.Store
	// name is the name of the table.
	name string
	// prefix is the prefix for all keys in the table.
	prefix uint32
}

// NewTableView creates a new view into a table in a TableStore.
func newTableView(
	base kvstore.Store,
	name string,
	prefix uint32) kvstore.Table {

	return &tableView{
		base:   base,
		name:   name,
		prefix: prefix,
	}
}

// Name returns the name of the table.
func (t *tableView) Name() string {
	return t.name
}

// Put adds a key-value pair to the table.
func (t *tableView) Put(key []byte, value []byte) error {
	k := t.TableKey(key)
	return t.base.Put(k, value)
}

// Get retrieves a value from the table.
func (t *tableView) Get(key []byte) ([]byte, error) {
	k := t.TableKey(key)
	return t.base.Get(k)
}

// Delete removes a key-value pair from the table.
func (t *tableView) Delete(key []byte) error {
	k := t.TableKey(key)
	return t.base.Delete(k)
}

// NewIterator creates a new iterator. Only keys prefixed with the given prefix will be iterated.
func (t *tableView) NewIterator(prefix []byte) (iterator.Iterator, error) {
	p := t.TableKey(prefix)
	return t.base.NewIterator(p)
}

// TODO it shouldn't be possible to shut down a table like this
// Shutdown shuts down the table.
func (t *tableView) Shutdown() error {
	return t.base.Shutdown()
}

// Destroy shuts down a table and deletes all data in it.
func (t *tableView) Destroy() error {
	return t.base.Destroy()
}

// tableBatch is a batch for a table in a TableStore.
type tableBatch struct {
	table kvstore.Table
	batch kvstore.Batch[[]byte]
}

// Put schedules a key-value pair to be added to the table.
func (t *tableBatch) Put(key []byte, value []byte) {
	k := t.table.TableKey(key)
	t.batch.Put(k, value)
}

// Delete schedules a key-value pair to be removed from the table.
func (t *tableBatch) Delete(key []byte) {
	k := t.table.TableKey(key)
	t.batch.Delete(k)
}

// Apply applies the batch to the table.
func (t *tableBatch) Apply() error {
	return t.batch.Apply()
}

// NewBatch creates a new batch for the table.
func (t *tableView) NewBatch() kvstore.Batch[[]byte] {
	return &tableBatch{
		table: t,
		batch: t.base.NewBatch(),
	}
}

// TableKey creates a key scoped to this table.
func (t *tableView) TableKey(key []byte) kvstore.TableKey {
	result := make([]byte, 4+len(key))
	binary.BigEndian.PutUint32(result, t.prefix)
	copy(result[4:], key)
	return result
}
