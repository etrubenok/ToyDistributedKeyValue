package kvstore

import (
	"errors"
	"github.com/linxGnu/grocksdb"
	"log"
)

type KVStore struct {
	db *grocksdb.DB
}

// NewKVStore creates a new KVStore instance
func NewKVStore(path string) (*KVStore, error) {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	// Options for RocksDB
	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	// Open RocksDB database (this creates or opens the DB in "path")
	var err error
	db, err := grocksdb.OpenDb(opts, path)
	if err != nil {
		log.Printf("Failed to open RocksDB: %v", err)
	}
	return &KVStore{db: db}, nil
}

// CloseDB Close the RocksDB database
func (store *KVStore) CloseDB() {
	store.db.Close()
}

// Set stores a key-value pair in RocksDB
func (store *KVStore) Set(key, value string) error {
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	return store.db.Put(wo, []byte(key), []byte(value))
}

// Get retrieves a value by key from RocksDB
func (store *KVStore) Get(key string) (string, error) {
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()

	value, err := store.db.Get(ro, []byte(key))
	if err != nil {
		return "", err
	}
	if !value.Exists() {
		return "", errors.New("key not found")
	}
	defer value.Free()

	return string(value.Data()), nil
}
