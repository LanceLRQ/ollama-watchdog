package utils

import (
	badger "github.com/dgraph-io/badger/v4"
)

func OpenBadgerDB(path string) (*badger.DB, error) {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return db, nil
}
