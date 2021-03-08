package main

import (
	"context"
	"google.golang.org/appengine/log"
	"io"

	pb "github.com/MikkelHJuul/ld/proto"

	"github.com/dgraph-io/badger/v3"
)

type ldService struct {
	db *badger.DB
}

func newServer() *ldService {
	db, err := badger.Open(badger.DefaultOptions("data/badger").WithInMemory(*mem))
	if err != nil {
		panic(err)
	}
	return &ldService{db: db}
}

func (l ldService) Create(ctx context.Context, value *pb.KeyValue) (*pb.CreateResponse, error) {
	err := l.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(value.Key), value.Value)
		return err
	})
	if err != nil {
		log.Warningf(ctx, "error while saving data to database ", err)
		return &pb.CreateResponse{Error: true}, err
	}
	return &pb.CreateResponse{}, nil
}

func (l ldService) CreateMany(server pb.Ld_CreateManyServer) error {
	txn := l.db.NewTransaction(true)
	for {
		create, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = server.Send(&pb.CreateResponse{Error: true})
			return err
		}
		if err := txn.Set([]byte(create.Key), create.Value); err == badger.ErrTxnTooBig {
			err = txn.Commit()
			if err != nil {
				return err //probably not?
			}
			txn = l.db.NewTransaction(true)
			err = txn.Set([]byte(create.Key), create.Value)
			if err != nil {
				return err //probably not?
			}
		}
		err = server.Send(&pb.CreateResponse{})
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

func (l ldService) Read(ctx context.Context, key *pb.Key) (*pb.KeyValue, error) {
	var value []byte
	err := l.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key.Key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Warningf(ctx, "error while fetching data from database ", err)
		return nil, err
	}
	return &pb.KeyValue{Key: key.Key, Value: value}, nil
}

func (l ldService) ReadMany(server pb.Ld_ReadManyServer) error {
	txn := l.db.NewTransaction(false)
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
		if err := l.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key.Key))
			if err != nil {
				return err
			}
			value, err = item.ValueCopy(nil)
			return err
		}); err == badger.ErrTxnTooBig {
			err = txn.Commit()
			if err != nil {
				return err //probably not?
			}
			txn = l.db.NewTransaction(false)
			item, err := txn.Get([]byte(key.Key))
			if err != nil {
				return err //probably not?
			}
			value, err = item.ValueCopy(nil)
			if err != nil {
				return err //probably not?
			}
		}
		err = server.Send(&pb.KeyValue{Key: key.Key, Value: value})
		if err != nil {
			return err
		}
	}
	return nil
}

func (l ldService) ReadRange(keyRange *pb.KeyRange, server pb.Ld_ReadRangeServer) error {
	panic("implement me using badger stream")
}

func (l ldService) Update(ctx context.Context, value *pb.KeyValue) (*pb.KeyValue, error) {
	kv, err := l.Read(ctx, &pb.Key{Key: value.Key}) // prefer more read-operations
	err = l.db.Update(func(txn *badger.Txn) error {
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
		log.Warningf(ctx, "error while saving data to database ", err)
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
		err = l.db.Update(func(txn *badger.Txn) error {
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
			log.Warningf(server.Context(), "error while saving data to database ", err)
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
	err = l.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key.Key))
	})
	if err != nil {
		log.Warningf(ctx, "error while deleting data in database ", err)
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
		err = l.db.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(key.Key))
		})
		if err != nil {
			log.Warningf(server.Context(), "error while deleting data in database ", err)
			return err
		}
		if err = server.Send(kv); err != nil {
			return err
		}
	}
	return nil
}

func (l ldService) DeleteRange(keyRange *pb.KeyRange, server pb.Ld_DeleteRangeServer) error {
	panic("implement me using badger stream")
}
