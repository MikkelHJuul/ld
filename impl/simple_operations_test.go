package impl

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	pb "github.com/MikkelHJuul/ld/proto"
)

type resp struct {
	KV *pb.KeyValue
	E  error
}

type kOp struct {
	Key  *pb.Key
	Resp resp
}

type kvOp struct {
	KV   *pb.KeyValue
	Resp resp
}

type aCase struct {
	name      string
	GetBefore []kOp
	Set       []kvOp
	Get       []kOp
	Del       []kOp
	GetAfter  []kOp
}

func TestSetGetDeleteSingles(t *testing.T) {
	lilly := &pb.KeyValue{Key: "hello", Value: []byte("Lilly")}
	ellis := &pb.KeyValue{Key: "hi", Value: []byte("Ellis")}
	cases := []aCase{
		{
			name: "test set, get and delete correct outcome",
			GetBefore: []kOp{
				{
					Key:  &pb.Key{Key: "hello"},
					Resp: resp{KV: nil},
				},
				{
					Key:  &pb.Key{Key: "hi"},
					Resp: resp{KV: nil},
				},
			},
			Set: []kvOp{
				{
					KV:   lilly,
					Resp: resp{},
				},
				{
					KV:   ellis,
					Resp: resp{},
				},
			},
			Get: []kOp{
				{
					Key:  &pb.Key{Key: "hello"},
					Resp: resp{KV: lilly},
				},
				{
					Key:  &pb.Key{Key: "hi"},
					Resp: resp{KV: ellis},
				},
				{ //assert twice
					Key:  &pb.Key{Key: "hi"},
					Resp: resp{KV: ellis},
				},
			},
			Del: []kOp{
				{
					Key:  &pb.Key{Key: "hello"},
					Resp: resp{KV: lilly},
				},
				{
					Key:  &pb.Key{Key: "hi"},
					Resp: resp{KV: ellis},
				},
				{ //assert second time it's gone
					Key:  &pb.Key{Key: "hi"},
					Resp: resp{KV: nil},
				},
			},
			GetAfter: []kOp{
				{
					Key:  &pb.Key{Key: "hello"},
					Resp: resp{KV: nil},
				},
				{
					Key:  &pb.Key{Key: "hi"},
					Resp: resp{KV: nil},
				},
			},
		},
		{
			name: "ingest empty-value object",
			Set: []kvOp{
				{
					KV:   &pb.KeyValue{Key: "empty"},
					Resp: resp{},
				},
			},
			Get: []kOp{
				{
					Key:  &pb.Key{Key: "empty"},
					Resp: resp{KV: &pb.KeyValue{Key: "empty"}},
				},
			},
			Del: []kOp{
				{
					Key:  &pb.Key{Key: "empty"},
					Resp: resp{KV: &pb.KeyValue{Key: "empty"}},
				},
			},
		},
	}
	for _, aCase := range cases {
		t.Run(aCase.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("tmp", "ldgsdtest")
			if err != nil {
				t.Fatal(err)
			}
			ld := NewServer(dir, false)
			ctx := context.Background()
			for _, op := range aCase.GetBefore {
				checkReturn(t, "GetBefore",
					func() (*pb.KeyValue, error) {
						return ld.Get(ctx, op.Key)
					}, op.Resp.KV)
			}
			for _, op := range aCase.Set {
				checkReturn(t, "Set",
					func() (*pb.KeyValue, error) {
						return ld.Set(ctx, op.KV)
					}, op.Resp.KV)
			}
			for _, op := range aCase.Get {
				checkReturn(t, "Get",
					func() (*pb.KeyValue, error) {
						return ld.Get(ctx, op.Key)
					}, op.Resp.KV)
			}
			for _, op := range aCase.Del {
				checkReturn(t, "Del",
					func() (*pb.KeyValue, error) {
						return ld.Delete(ctx, op.Key)
					}, op.Resp.KV)
			}
			for _, op := range aCase.GetAfter {
				checkReturn(t, "GetAfter",
					func() (*pb.KeyValue, error) {
						return ld.Get(ctx, op.Key)
					}, op.Resp.KV)
			}
		})
	}
}

func checkReturn(t *testing.T, name string, methd func() (*pb.KeyValue, error), equalsObj *pb.KeyValue) {
	kv, err := methd()
	if err != nil {
		t.Errorf("unexpected error in %s: %v", name, err)
	}
	if !reflect.DeepEqual(kv, equalsObj) {
		t.Errorf("response in %s returned unexpected response: %v", name, kv)
	}
}
