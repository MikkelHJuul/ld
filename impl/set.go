package impl

import (
	"context"
	pb "github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	"io"
	"log"
)

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
	done := make(chan int)
	out := l.setManyGenerator(in)
	go func() {
		for kv := range out {
			if err := server.Send(kv); err != nil {
				log.Print(err)
			}
		}
		done <- 1
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
	<-done
	return nil
}
