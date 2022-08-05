package main

import (
	"context"

	v1 "github.com/MikkelHJuul/ld/gen/ld/v1"
	ldv1 "github.com/MikkelHJuul/ld/gen/ld/v1/ldv1connect"
	connect_go "github.com/bufbuild/connect-go"
)

type ldKv struct {
	db kvDatabase
}

var _ ldv1.LdServiceHandler = (*ldKv)(nil)

func checkCtx(ctx context.Context) func() error {
	return func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	}
}

// Delete implements ldv1connect.LdServiceHandler
func (lk *ldKv) Delete(_ context.Context, req *connect_go.Request[v1.DeleteRequest]) (*connect_go.Response[v1.DeleteResponse], error) {
	val, err := lk.db.delete(req.Msg.K.Key)
	if err != nil {
		return nil, err
	}
	return connect_go.NewResponse(&v1.DeleteResponse{Kv: &v1.KeyValue{Key: req.Msg.K.Key, Value: val}}), nil
}

// Get implements ldv1connect.LdServiceHandler
func (lk *ldKv) Get(_ context.Context, req *connect_go.Request[v1.GetRequest]) (*connect_go.Response[v1.GetResponse], error) {
	val, err := lk.db.get(req.Msg.K.Key)
	if err != nil {
		return nil, err
	}
	return connect_go.NewResponse(&v1.GetResponse{Kv: &v1.KeyValue{Key: req.Msg.K.Key, Value: val}}), nil
}

// Set implements ldv1connect.LdServiceHandler
func (lk *ldKv) Set(ctx context.Context, req *connect_go.Request[v1.SetRequest]) (*connect_go.Response[v1.SetResponse], error) {
	val, err := lk.db.set(kv{key: req.Msg.Kv.Key, value: req.Msg.Kv.Value})
	if err != nil {
		return nil, err
	}
	return connect_go.NewResponse(&v1.SetResponse{Kv: &v1.KeyValue{Key: val.key, Value: val.value}}), nil
}
