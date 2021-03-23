package impl

import (
	"context"
	"io"

	pb "github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

// Delete implements the RPC method of proto.LdServer.
// returns the deleted KeyValue or nil for no such key
func (l ldService) Delete(_ context.Context, key *pb.Key) (*pb.KeyValue, error) {
	var value []byte
	err := l.DB.Update(func(txn *badger.Txn) (err error) {
		value, err = readSingleFromKey(txn, key)
		if err != nil {
			return
		}
		return txn.Delete([]byte(key.Key))
	})
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		log.Errorf("error while deleting data in database: %v ", err)
		return nil, err
	}
	return &pb.KeyValue{Key: key.Key, Value: value}, nil
}

// DeleteMany implements the relevant RPC of proto.LdServer
func (l ldService) DeleteMany(server pb.Ld_DeleteManyServer) error {
	out := make(chan *pb.KeyValue)
	keys := make(chan *pb.Key)
	done := make(chan int)

	//go routine that just sends!
	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Warn(err)
			}
		}
		done <- 1
	}()

	go func() {
		txn := l.DB.NewTransaction(true)
		defer txn.Commit()
		for k := range keys {
			var value []byte
			value, err := readSingleFromKey(txn, k)
			if err == badger.ErrKeyNotFound {
				out <- nil
				continue
			}
			if err != nil {
				log.Info("error reading before delete", err)
				continue
			}
			if err = txn.Delete([]byte(k.Key)); err == badger.ErrTxnTooBig {
				err = txn.Commit()
				if err != nil {
					log.Warn(err)
				}
				if err = txn.Delete([]byte(k.Key)); err != nil {
					log.Warn(err)
				}
			}
			if err != nil {
				log.Info("could not delete record", err)
			}
			out <- &pb.KeyValue{Key: k.Key, Value: value}
		}
		close(out)
	}()

	for {
		key, err := server.Recv()
		if err == io.EOF {
			close(keys)
			break
		}
		if err != nil {
			log.Warn("error from Ld_DeleteManyServer", err)
			return err
		}
		keys <- key
	}
	<-done
	return nil
}

// DeleteRange implements delete functionality for a query object to fulfill interface method from proto.LdServer
func (l ldService) DeleteRange(keyRange *pb.KeyRange, server pb.Ld_DeleteRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		log.Debugf("Could not compile matcher from patter, %v: %v", keyRange.Pattern, err)
		return err
	}
	chKeyMatches := make(chan *pb.Key)
	out := make(chan *pb.KeyValue)
	done := make(chan int)

	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Info(err)
			}
		}
		done <- 1
	}()

	go func() {
		txn := l.DB.NewTransaction(true)
		defer txn.Commit()
		for key := range chKeyMatches {
			value, err := readSingleFromKey(txn, key)
			if err == badger.ErrKeyNotFound {
				out <- nil
				err = nil
			}
			if err != nil {
				return
			}
			err = txn.Delete([]byte(key.Key))
			if err != nil {
				out <- nil
				log.Info("error when deleting record", err)
				continue
			}
			out <- &pb.KeyValue{Key: key.Key, Value: value}
		}
		close(out)
	}()

	if err = l.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		iter := keyRangeIterator(it, keyRange)
		for iter.Rewind(); iter.Valid(); iter.Next() {
			k := iter.Item().Key()
			if matcher.Match(k) {
				chKeyMatches <- &pb.Key{Key: string(k)}
			}
		}
		return nil
	}); err != nil {
		log.Warn("error finding keys", err)
		return err
	}
	close(chKeyMatches)
	<-done
	return nil
}
