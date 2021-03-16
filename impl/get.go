package impl

import (
	"context"
	pb "github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	"io"
	"log"
	"sync"
)

func (l ldService) Get(_ context.Context, key *pb.Key) (*pb.KeyValue, error) {
	var value []byte
	err := l.DB.View(func(txn *badger.Txn) (err error) {
		value, err = readSingleFromKey(txn, key)
		return
	})
	if err != nil {
		log.Printf("error while fetching data from database: %v", err)
		return nil, err
	}
	return &pb.KeyValue{Key: key.Key, Value: value}, nil
}

func (l ldService) GetMany(server pb.Ld_GetManyServer) error {
	txn := l.DB.NewTransaction(false)
	defer txn.Commit()
	out := make(chan *pb.KeyValue)
	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
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
			return err
		}
		wg.Add(1)
		go l.sendKeyWith(out, txn, wg, key)
	}
	return nil
}

func (l ldService) GetRange(keyRange *pb.KeyRange, server pb.Ld_GetRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		log.Printf("Could not compile matcher from patter, %v: %v", keyRange.Pattern, err)
		return err
	}
	chKeyMatches := make(chan *pb.Key)
	out := make(chan *pb.KeyValue)
	done := make(chan int)

	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
		done <- 1
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
		log.Print("error finding keys", err)
		return err
	}
	close(chKeyMatches)
	<-done
	return nil
}
