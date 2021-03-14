package impl

import (
	"context"
	"io"
	"log"

	pb "github.com/MikkelHJuul/ld/proto"

	"github.com/dgraph-io/badger/v3"
)

type ldService struct {
	*badger.DB
}

func NewServer(inmem bool) *ldService {
	db, err := badger.Open(badger.DefaultOptions("ld_badger").WithInMemory(inmem))
	if err != nil {
		log.Fatal(err)
	}
	return &ldService{db}
}

func (l ldService) Set(_ context.Context, value *pb.KeyValue) (*pb.KeyValue, error) {
	err := l.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(value.Key), value.Value)
	})
	if err != nil {
		log.Printf("error while saving data to database: %v ", err)
		return value, err
	}
	return nil, nil
}

func (l ldService) SetMany(server pb.Ld_SetManyServer) error {
	in := make(chan *pb.KeyValue)
	out := l.setManyGenerator(in)
	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
	}()
	for {
		create, err := server.Recv()
		if err == io.EOF {
			close(in)
			break
		}
		if err != nil {
			return err
		}
		in <- create
	}
	return nil
}

func (l ldService) setManyGenerator(in chan *pb.KeyValue) chan *pb.KeyValue {
	out := make(chan *pb.KeyValue)
	go func() {
		txn := l.DB.NewTransaction(true)
		for create := range in {
			if err := txn.Set([]byte(create.Key), create.Value); err == badger.ErrTxnTooBig {
				err = txn.Commit()
				if err != nil {
					log.Print("error when committing transaction in goroutine", err) //probably not?
				}
				txn = l.DB.NewTransaction(true)
				err = txn.Set([]byte(create.Key), create.Value)
				if err != nil {
					log.Print("error when setting ", err) //probably not?
				}
			}
		}
		close(out)
		if err := txn.Commit(); err != nil {
			log.Print("error when committing transaction in goroutine", err) //probably not?
		}
	}()
	return out
}

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
	for {
		key, err := server.Recv()
		if err == io.EOF {
			close(out)
			break
		}
		if err != nil {
			return err
		}
		go func() {
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
		}()
	}
	return nil
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

//fix after this!

func (l ldService) GetRange(keyRange *pb.KeyRange, server pb.Ld_GetRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		log.Printf("Could not compile matcher from patter, %v: %v", keyRange.Pattern, err)
		return err
	}
	chMatches := make(chan []byte)

	go func() {
		select {
		case key := <-chMatches:
			txn := l.DB.NewTransaction(false)
			defer txn.Discard()
			item, err := txn.Get(key)
			if err == badger.ErrTxnTooBig {
				if err = txn.Commit(); err != nil {
					log.Print("error while commit too large transaction early", err)
				}
				txn = l.DB.NewTransaction(false)
				if item, err = txn.Get(key); err != nil {
					log.Print("could not fetch value for key", err)
				}
			}
			if item != nil {
				itemVal, err := item.ValueCopy(nil)
				if err != nil {
					log.Print("could not copy value of the item", err)
				}
				if err = server.Send(&pb.KeyValue{Key: string(key), Value: itemVal}); err != nil {
					log.Print("error sending the KeyValue object", err)
				}
			}
			err = txn.Commit()
			if err != nil {
				log.Printf("could not commit transaction")
			}
		}
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
				chMatches <- k
			}
		}
		close(chMatches)
		return nil
	}); err != nil {
		log.Print("error finding keys", err)
		return err
	}
	return nil
}

func (l ldService) Delete(ctx context.Context, key *pb.Key) (*pb.KeyValue, error) {
	kv, err := l.Get(ctx, key) // prefer more read-operations
	if err != nil || kv.Value == nil {
		return nil, err
	}
	err = l.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key.Key))
	})
	if err != nil {
		log.Printf("error while deleting data in database: %v ", err)
		return nil, err
	}
	return kv, nil
}

func (l ldService) DeleteMany(server pb.Ld_DeleteManyServer) error {
	for {
		key, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		kv, err := l.Get(server.Context(), key) // prefer more read-operations
		if err != nil {
			return err
		}
		if kv.Value == nil {
			if err = server.Send(&pb.KeyValue{Key: key.Key}); err != nil {
				return err
			}
			continue
		}
		err = l.DB.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(key.Key))
		})
		if err != nil {
			log.Printf("error while deleting data in database: %v", err)
			return err
		}
		if err = server.Send(kv); err != nil {
			return err
		}
	}
	return nil
}

func (l ldService) DeleteRange(keyRange *pb.KeyRange, server pb.Ld_DeleteRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		return err
	}
	chMatches := make(chan []byte)
	go l.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		iter := keyRangeIterator(it, keyRange)
		for iter.Rewind(); iter.Valid(); iter.Next() {
			k := iter.Item().Key()
			if matcher.Match(k) {
				chMatches <- k
			}
		}
		close(chMatches)
		return nil
	})
	txn := l.DB.NewTransaction(true)
	defer txn.Discard()
	for key := range chMatches {
		item, err := txn.Get(key)
		if err == badger.ErrTxnTooBig {
			if err = txn.Commit(); err != nil {
				return err
			}
			txn = l.DB.NewTransaction(true)
			item, err = txn.Get(key)
			if err != nil {
				return err
			}
		}
		if err := txn.Delete(key); err == badger.ErrTxnTooBig {
			if err = txn.Commit(); err != nil {
				return err
			}
			txn = l.DB.NewTransaction(true)
			if err = txn.Delete(key); err != nil {
				return err
			}
		}
		var itemVal []byte
		if item != nil {
			itemVal, err = item.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		if err = server.Send(&pb.KeyValue{Key: string(key), Value: itemVal}); err != nil {
			return err
		}
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}
