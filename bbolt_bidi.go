package main

import (
	"context"

	v1 "github.com/MikkelHJuul/ld/gen/ld/v1"
	connect_go "github.com/bufbuild/connect-go"
)

type sendRecvBidiStream[req, res any] struct {
	stream   *connect_go.BidiStream[req, res]
	check    func() error
	kvMeth   func(*req) kv
	respMeth func(kv) *res
}

var _ sender = (*sendRecvBidiStream[any, any])(nil)
var _ receiver = (*sendRecvBidiStream[any, any])(nil)

func (s *sendRecvBidiStream[req, res]) Send(keyVal kv) error {
	if err := s.stream.Send(s.respMeth(keyVal)); err != nil {
		return err
	}
	return nil
}

// Receive implements receiver
func (s *sendRecvBidiStream[req, res]) Receive() kv {
	if err := s.check(); err != nil {
		return kv{error: err}
	}

	streamReq, err := s.stream.Receive()
	if err != nil {
		return kv{error: err}
	}

	return s.kvMeth(streamReq)
}

// DeleteMany implements ldv1connect.LdServiceHandler
func (lk *ldKv) DeleteMany(ctx context.Context, stream *connect_go.BidiStream[v1.DeleteManyRequest, v1.DeleteManyResponse]) error {
	streamer := &sendRecvBidiStream[v1.DeleteManyRequest, v1.DeleteManyResponse]{
		stream: stream,
		check:  checkCtx(ctx),
		kvMeth: func(k *v1.DeleteManyRequest) kv {
			return kv{key: k.GetK().Key}
		},
		respMeth: func(keyVal kv) *v1.DeleteManyResponse {
			return &v1.DeleteManyResponse{Kv: &v1.KeyValue{Key: keyVal.key, Value: keyVal.value}}
		},
	}
	return lk.db.deleteMany(streamer, streamer)
}

// GetMany implements ldv1connect.LdServiceHandler
func (lk *ldKv) GetMany(ctx context.Context, stream *connect_go.BidiStream[v1.GetManyRequest, v1.GetManyResponse]) error {
	streamer := &sendRecvBidiStream[v1.GetManyRequest, v1.GetManyResponse]{
		stream: stream,
		check:  checkCtx(ctx),
		kvMeth: func(k *v1.GetManyRequest) kv {
			return kv{key: k.GetK().Key}
		},
		respMeth: func(keyVal kv) *v1.GetManyResponse {
			return &v1.GetManyResponse{Kv: &v1.KeyValue{Key: keyVal.key, Value: keyVal.value}}
		},
	}
	return lk.db.getMany(streamer, streamer)
}

// SetMany implements ldv1connect.LdServiceHandler
func (lk *ldKv) SetMany(ctx context.Context, stream *connect_go.BidiStream[v1.SetManyRequest, v1.SetManyResponse]) error {
	streamer := &sendRecvBidiStream[v1.SetManyRequest, v1.SetManyResponse]{
		stream: stream,
		check:  checkCtx(ctx),
		kvMeth: func(k *v1.SetManyRequest) kv {
			return kv{key: k.Kv.Key, value: k.Kv.Value}
		},
		respMeth: func(keyVal kv) *v1.SetManyResponse {
			return &v1.SetManyResponse{Kv: &v1.KeyValue{Key: keyVal.key, Value: keyVal.value}}
		},
	}
	return lk.db.getMany(streamer, streamer)
}
