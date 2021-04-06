package impl

import (
	"context"
	"io"
	"sync"

	pb "github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

// Get implements RPC method Get, returns nil/empty message for no such key.
func (l ldService) Get(_ context.Context, key *pb.Key) (*pb.KeyValue, error) {
	var value []byte
	err := l.DB.View(func(txn *badger.Txn) (err error) {
		value, err = readSingleFromKey(txn, key)
		return
	})
	return decideOutcome(err, key.Key, value)
}

// GetMany implements RPC stream method of the same name from LdServer
func (l ldService) GetMany(server pb.Ld_GetManyServer) error {
	txn := l.DB.NewTransaction(false)
	defer txn.Commit()
	out := make(chan *pb.KeyValue)
	ctx, ccl := context.WithCancel(context.Background())
	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Warn(err)
			}
		}
		ccl()
	}()
	wg := &sync.WaitGroup{}
	for {
		key, err := server.Recv()
		if err == io.EOF {
			wg.Wait()
			close(out)
			break
		}
		if err != nil {
			log.Warn("error from Ld_GetManyServer", err)
			return err
		}
		wg.Add(1)
		go l.sendKeyWith(out, txn, wg, key)
	}
	<-ctx.Done()
	return nil
}

// GetRange implements RPC query method from proto.LdServer
func (l ldService) GetRange(keyRange *pb.KeyRange, server pb.Ld_GetRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		log.Debugf("Could not compile matcher from patter, %v: %v", keyRange.Pattern, err)
		return err
	}
	chKeyMatches := make(chan *pb.Key)
	out := make(chan *pb.KeyValue)
	ctx, ccl := context.WithCancel(context.Background())

	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Warn(err)
			}
		}
		ccl()
	}()

	go func() {
		wg := &sync.WaitGroup{}
		txn := l.DB.NewTransaction(false)
		defer txn.Commit()
		for key := range chKeyMatches {
			wg.Add(1)
			go l.sendKeyWith(out, txn, wg, key)
		}
		wg.Wait()
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
	<-ctx.Done()
	return nil
}
