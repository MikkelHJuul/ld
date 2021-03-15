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
		go l.sendKeyWith(out, txn, key)
	}
	return nil
}

func (l ldService) sendKeyWith(out chan *pb.KeyValue, txn *badger.Txn, key *pb.Key) {
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

func (l ldService) GetRange(keyRange *pb.KeyRange, server pb.Ld_GetRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		log.Printf("Could not compile matcher from patter, %v: %v", keyRange.Pattern, err)
		return err
	}
	chKeyMatches := make(chan *pb.Key)
	out := make(chan *pb.KeyValue)

	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
	}()

	go func() {
		txn := l.DB.NewTransaction(false)
		defer txn.Commit()
		for key := range chKeyMatches {
			go l.sendKeyWith(out, txn, key)
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
		close(chKeyMatches)
		return nil
	}); err != nil {
		log.Print("error finding keys", err)
		return err
	}
	return nil
}

func (l ldService) Delete(_ context.Context, key *pb.Key) (*pb.KeyValue, error) {
	var value []byte
	err := l.DB.Update(func(txn *badger.Txn) (err error) {
		value, err = readSingleFromKey(txn, key)
		if err != nil {
			return
		}
		return txn.Delete([]byte(key.Key))
	})
	if err != nil {
		log.Printf("error while deleting data in database: %v ", err)
		return nil, err
	}
	return &pb.KeyValue{Key: key.Key, Value: value}, nil
}

func (l ldService) DeleteMany(server pb.Ld_DeleteManyServer) error {
	out := make(chan *pb.KeyValue)
	keys := make(chan *pb.Key)

	//go routine that just sends!
	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
	}()

	go func() {
		txn := l.DB.NewTransaction(true)
		defer txn.Commit()
		for k := range keys {
			var value []byte
			value, err := readSingleFromKey(txn, k)
			//log...
			if err == badger.ErrKeyNotFound {
				out <- &pb.KeyValue{}
				continue //skip log -- implement debug log
			}
			if err != nil {
				log.Print(err)
				continue
			}
			if err = txn.Delete([]byte(k.Key)); err == badger.ErrTxnTooBig {
				err = txn.Commit()
				if err != nil {
					log.Print(err) //uncommitted read-transaction... hope it is fine
				}
				if err = txn.Delete([]byte(k.Key)); err != nil {
					log.Print(err)
				}
			}
			if err != nil {
				log.Print("could not delete record", err)
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
			return err
		}
		keys <- key
	}
	return nil
}

func (l ldService) DeleteRange(keyRange *pb.KeyRange, server pb.Ld_DeleteRangeServer) error {
	matcher, err := NewMatcher(keyRange.Pattern)
	if err != nil {
		log.Printf("Could not compile matcher from patter, %v: %v", keyRange.Pattern, err)
		return err
	}
	chKeyMatches := make(chan *pb.Key)
	out := make(chan *pb.KeyValue)

	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
	}()

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
			err = txn.Delete([]byte(key.Key))
			if err != nil {
				out <- &pb.KeyValue{}
				log.Print("error when deleting record", err)
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
		close(chKeyMatches)
		return nil
	}); err != nil {
		log.Print("error finding keys", err)
		return err
	}
	return nil
}
