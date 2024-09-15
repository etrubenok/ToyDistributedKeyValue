package main

import (
	"fmt"
	"github.com/linxGnu/grocksdb"
)

func main() {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	db, err := grocksdb.OpenDb(opts, "./testdb")
	// check for error
	if err != nil {
		// Panic with the message saying that the database could not be opened
		panic(fmt.Errorf("failed to open database: %v", err))
	}

	ro := grocksdb.NewDefaultReadOptions()
	wo := grocksdb.NewDefaultWriteOptions()

	err = db.Put(wo, []byte("key"), []byte("value"))
	// check for error
	if err != nil {
		panic(fmt.Errorf("failed to write to database: %v", err))
	}

	value, err := db.Get(ro, []byte("key"))
	// check for error
	if err != nil {
		panic(fmt.Errorf("failed to read from database: %v", err))
	}
	defer value.Free()

	println(string(value.Data()))

	err = db.Delete(wo, []byte("key"))
	// check for error
	if err != nil {
		panic(fmt.Errorf("failed to delete from database: %v", err))
	}
}
