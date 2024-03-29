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
		return txn.Delete(key.Key)
	})
	if err == badger.ErrKeyNotFound {
		return &pb.KeyValue{}, nil
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
	ctx, ccl := context.WithCancel(context.Background())

	go sendCancel(server, out, ccl)

	go func() {
		txn := l.DB.NewTransaction(true)
		defer txn.Commit()
		for k := range keys {
			var value []byte
			value, err := readSingleFromKey(txn, k)
			if err == badger.ErrKeyNotFound {
				out <- &pb.KeyValue{}
				continue
			}
			if err != nil {
				log.Info("error reading before delete", err)
				break
			}
			err = l.deleteTransaction(txn, k)
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
	<-ctx.Done()
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
	ctx, ccl := context.WithCancel(context.Background())

	go sendCancel(server, out, ccl)

	go func() {
		txn := l.DB.NewTransaction(true)
		defer txn.Commit()
		for key := range chKeyMatches {
			value, err := readSingleFromKey(txn, key)
			if err == badger.ErrKeyNotFound {
				out <- &pb.KeyValue{}
				err = nil
			}
			if err != nil {
				return
			}
			err = l.deleteTransaction(txn, key)
			if err != nil {
				out <- nil
				log.Info("error when deleting record", err)
				break
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
				chKeyMatches <- &pb.Key{Key: k}
			}
		}
		return nil
	}); err != nil {
		log.Warn("error finding keys", err)
		return err
	}
	close(chKeyMatches)
	<-ctx.Done()
	return nil
}
