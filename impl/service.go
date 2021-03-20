package impl

import (
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
func NewServer(dbLocation string, inmem bool) *ldService {
	if inmem {
		dbLocation = ""
	}
	db, err := badger.Open(badger.DefaultOptions(dbLocation).WithInMemory(inmem))
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
			log.Warn(err) //uncommitted read-transaction... hope it is fine
		}
		if err = sendKeyValue(out, txn, key); err != nil {
			log.Error("could not finish transaction after failure")
		}
	}
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
			return nil, nil
		}
		log.Warn("error in transaction", err)
		return nil, err
	}
	return &pb.KeyValue{Key: key, Value: value}, nil
}
