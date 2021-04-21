package impl

import (
	"github.com/MikkelHJuul/bIter"
	pb "github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"

	"sync"
)

type ldService struct {
	*badger.DB
}

// NewServer opens and returns a badger.DB facade
// that implements the proto interface proto.LdServer.
func NewServer(dbInitOptions ...func(*badger.Options)) (*ldService, error) {
	o := badger.DefaultOptions("ld_badger")
	for _, opt := range dbInitOptions {
		opt(&o)
	}
	db, err := badger.Open(o)
	if err != nil {
		return nil, err
	}
	return &ldService{db}, nil
}

func (l ldService) sendKeyWith(out chan *pb.KeyValue, txn *badger.Txn, wg *sync.WaitGroup, key *pb.Key) {
	defer wg.Done()
	err := l.handleKeyTransaction(txn, key, false, func(txn *badger.Txn, key *pb.Key) error {
		return sendKeyValue(out, txn, key)
	})
	if err != nil {
		log.Error(err)
	}
}

func (l ldService) deleteTransaction(txn *badger.Txn, key *pb.Key) error {
	return l.handleKeyTransaction(txn, key, true, func(txn *badger.Txn, key *pb.Key) error {
		return txn.Delete([]byte(key.Key))
	})
}

func (l ldService) handleKeyTransaction(txn *badger.Txn, key *pb.Key, update bool, meth func(txn *badger.Txn, key *pb.Key) error) error {
	err := meth(txn, key)
	if err == badger.ErrTxnTooBig {
		err = txn.Commit()
		if err != nil {
			log.Warn(err) //uncommitted read-transaction... hope it is fine
		}
		txn = l.DB.NewTransaction(update)
		if err = meth(txn, key); err != nil {
			log.Error("could not finish transaction after failure")
		}
	}
	return err
}

func sendKeyValue(out chan *pb.KeyValue, txn *badger.Txn, key *pb.Key) error {
	var value []byte
	value, err := readSingleFromKey(txn, key)
	kv, err := decideOutcome(err, key.Key, value)
	out <- kv
	return err
}

func decideOutcome(err error, key string, value []byte) (*pb.KeyValue, error) {
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return &pb.KeyValue{}, nil
		}
		log.Warn("error in transaction", err)
		return nil, err
	}
	return &pb.KeyValue{Key: key, Value: value}, nil
}

func keyRangeIterator(it *badger.Iterator, keyRange *pb.KeyRange) bIter.Iterator {
	return bIter.KeyRangeIterator(it, []byte(keyRange.Prefix), []byte(keyRange.From), []byte(keyRange.To))
}

func readSingleFromKey(txn *badger.Txn, key *pb.Key) (value []byte, err error) {
	return readSingle(txn, []byte(key.Key))
}

func readSingle(txn *badger.Txn, key []byte) (value []byte, err error) {
	value = nil
	item, err := txn.Get(key)
	if err == nil {
		return item.ValueCopy(nil)
	}
	return
}
