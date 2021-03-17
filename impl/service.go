package impl

import (
	pb "github.com/MikkelHJuul/ld/proto"
	"log"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

type ldService struct {
	*badger.DB
}

// NewServer opens and returns a badger.DB facade
// that implements the proto interface proto.LdServer.
func NewServer(inmem bool) *ldService {
	db, err := badger.Open(badger.DefaultOptions("ld_badger").WithInMemory(inmem))
	if err != nil {
		log.Fatal(err)
	}
	return &ldService{db}
}

func (l ldService) sendKeyWith(out chan *pb.KeyValue, txn *badger.Txn, wg *sync.WaitGroup, key *pb.Key) {
	defer wg.Done()
	err := sendKeyValue(out, txn, key)
	if err == badger.ErrTxnTooBig {
		err = txn.Commit()
		if err != nil {
			log.Print(err) //uncommitted read-transaction... hope it is fine
		}
		if err = sendKeyValue(out, txn, key); err != nil {
			log.Print("could not finish transaction after failure")
		}
	}
}

func sendKeyValue(out chan *pb.KeyValue, txn *badger.Txn, key *pb.Key) error {
	var value []byte
	value, err := readSingleFromKey(txn, key)
	if err == badger.ErrKeyNotFound {
		out <- &pb.KeyValue{}
		err = nil
	}
	out <- &pb.KeyValue{Key: key.Key, Value: value}
	return err
}
