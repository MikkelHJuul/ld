package impl

import (
	"context"
	"io"

	pb "github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

// Set implements method Set from proto.LdServer. returns nothing for succes, the value and error for any error.
func (l ldService) Set(_ context.Context, value *pb.KeyValue) (*pb.KeyValue, error) {
	err := l.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(value.Key), value.Value)
	})
	if err != nil {
		log.Errorf("error while saving data to database: %v ", err)
		return value, err
	}
	return &pb.KeyValue{}, nil
}

// SetMany implements method SetMany from proto.LdServer.
func (l ldService) SetMany(server pb.Ld_SetManyServer) error {
	in := make(chan *pb.KeyValue)
	ctx, ccl := context.WithCancel(context.Background())
	out := l.setManyGenerator(in)
	go sendCancel(server, out, ccl)
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
	<-ctx.Done()
	return nil
}

func (l ldService) setManyGenerator(in chan *pb.KeyValue) chan *pb.KeyValue {
	out := make(chan *pb.KeyValue)
	go func() {
		txn := l.DB.NewTransaction(true)
		for create := range in {
			err := l.handleKeyTransaction(txn, &pb.Key{Key: create.Key}, true, func(txn *badger.Txn, key *pb.Key) error {
				return txn.Set([]byte(create.Key), create.Value)
			})
			if err != nil {
				out <- create
			} else {
				out <- &pb.KeyValue{}
			}
		}
		close(out)
		if err := txn.Commit(); err != nil {
			log.Warn("error when committing transaction in goroutine", err) //probably not?
		}
	}()
	return out
}
