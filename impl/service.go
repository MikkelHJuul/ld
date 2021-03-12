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

func (l ldService) Create(_ context.Context, value *pb.KeyValue) (*pb.KeyValue, error) {
	err := l.DB.View(func(txn *badger.Txn) (err error) {
		_, err = readSingleItem(txn, []byte(value.Key))
		return
	})
	if err != badger.ErrKeyNotFound {
		return value, err
	}
	err = l.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(value.Key), value.Value)
	})
	if err != nil {
		log.Printf("error while saving data to database: %v ", err)
		return value, err
	}
	return nil, nil
}

func (l ldService) CreateMany(server pb.Ld_CreateManyServer) error {
	txn := l.DB.NewTransaction(true)
	for {
		create, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = server.Send(create)
			return err
		}
		err = l.DB.View(func(txn *badger.Txn) (err error) {
			_, err = readSingleItem(txn, []byte(create.Key))
			return
		})
		if err != badger.ErrKeyNotFound {
			//there is already a record with the given key
			if err = server.Send(create); err != nil {
				return err
			}
			continue
		}
		log.Printf("got message, key: %s", create.Key)
		if err := txn.Set([]byte(create.Key), create.Value); err == badger.ErrTxnTooBig {
			err = txn.Commit()
			if err != nil {
				return err //probably not?
			}
			txn = l.DB.NewTransaction(true)
			err = txn.Set([]byte(create.Key), create.Value)
			if err != nil {
				return err //probably not?
			}
		}
		err = server.Send(&pb.KeyValue{})
		if err != nil {
			return err //probably not?
		}
	}
	err := txn.Commit()
	if err != nil {
		return err //probably not?
	}
	return nil
}

func (l ldService) Read(_ context.Context, key *pb.Key) (*pb.KeyValue, error) {
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

func (l ldService) ReadMany(server pb.Ld_ReadManyServer) error {
	txn := l.DB.NewTransaction(false)
	defer txn.Commit()
	for {
		key, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		var value []byte
		if err := l.DB.View(func(txn *badger.Txn) (err error) {
			value, err = readSingleFromKey(txn, key)
			return
		}); err == badger.ErrTxnTooBig {
			err = txn.Commit()
			if err != nil {
				log.Print(err) //uncommitted read-transaction... hope it is fine
			}
			txn = l.DB.NewTransaction(false)
			value, err = readSingleFromKey(txn, key)
			if err != nil {
				log.Print("could not read transaction after failure")
			}
		} else if err == badger.ErrKeyNotFound {
			if err = server.Send(&pb.KeyValue{}); err != nil { //todo decide if null item is the way
				return err
			}
		}
		if err = server.Send(&pb.KeyValue{Key: key.Key, Value: value}); err != nil {
			return err
		}
	}
	return nil
}

func (l ldService) ReadRange(keyRange *pb.KeyRange, server pb.Ld_ReadRangeServer) error {
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
		return nil
	}); err != nil {
		log.Print("error finding keys", err)
		return err
	}
	return nil
}

func (l ldService) Update(ctx context.Context, value *pb.KeyValue) (*pb.KeyValue, error) {
	kv, err := l.Read(ctx, &pb.Key{Key: value.Key}) // prefer more read-operations
	err = l.DB.Update(func(txn *badger.Txn) error {
		if kv.Value != nil {
			err := txn.Delete([]byte(value.Key))
			if err != nil {
				return err
			}
		} else {
			kv = value
		}
		err = txn.Set([]byte(value.Key), value.Value)
		return err
	})
	if err != nil {
		log.Printf("error while saving data to database: %v ", err)
		return nil, err
	}
	return kv, nil
}

func (l ldService) UpdateMany(server pb.Ld_UpdateManyServer) error {
	for {
		keyValue, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		kv, err := l.Read(server.Context(), &pb.Key{Key: keyValue.Key}) // prefer more read-operations
		err = l.DB.Update(func(txn *badger.Txn) error {
			if kv.Value != nil {
				err := txn.Delete([]byte(keyValue.Key))
				if err != nil {
					return err
				}
			} else {
				kv = keyValue
			}
			err = txn.Set([]byte(keyValue.Key), keyValue.Value)
			return err
		})
		if err != nil {
			log.Printf("error while saving data to database: %v ", err)
			return err
		}
		if err = server.Send(kv); err != nil {
			return err
		}
	}
	return nil
}

func (l ldService) Delete(ctx context.Context, key *pb.Key) (*pb.KeyValue, error) {
	kv, err := l.Read(ctx, key) // prefer more read-operations
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
		kv, err := l.Read(server.Context(), key) // prefer more read-operations
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
	chMatches := make(chan []byte, 1000) // channel size?
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
		if item != nil {
			itemVal, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if err = server.Send(&pb.KeyValue{Key: string(key), Value: itemVal}); err != nil {
				return err
			}
		}
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}
