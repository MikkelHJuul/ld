package impl

import (
	"github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
)

func readSingleFromKey(txn *badger.Txn, key *proto.Key) (value []byte, err error) {
	return readSingle(txn, []byte(key.Key))
}

func readSingle(txn *badger.Txn, key []byte) (value []byte, err error) {
	value = nil
	if item, err := readSingleItem(txn, key); err == nil {
		return item.ValueCopy(nil)
	}
	return
}

func readSingleItem(txn *badger.Txn, key []byte) (*badger.Item, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	return item, nil
}
