package impl

import (
	"fmt"
	"github.com/MikkelHJuul/ld/proto"
	"google.golang.org/grpc"
	"io"
)

type testBidiServer struct {
	grpc.ServerStream
	send    []*proto.KeyValue
	receive []*proto.KeyValue
	idx     int
}

type testBidiKeyServer struct {
	testBidiServer
	send []*proto.Key
}

func NewTestKeyServer(ks []*proto.Key) *testBidiKeyServer {
	return &testBidiKeyServer{
		send: ks,
		testBidiServer: testBidiServer{
			receive: make([]*proto.KeyValue, 0),
			idx:     0,
		},
	}
}

func NewTestServer(kvs []*proto.KeyValue) *testBidiServer {
	return &testBidiServer{
		send:    kvs,
		receive: make([]*proto.KeyValue, 0),
		idx:     0,
	}
}

func (s *testBidiKeyServer) Recv() (*proto.Key, error) {
	if len(s.send) == s.idx {
		return nil, io.EOF
	}
	val := s.send[s.idx]
	s.idx++
	return val, nil
}

func (s *testBidiServer) Recv() (*proto.KeyValue, error) {
	if len(s.send) == s.idx {
		return nil, io.EOF
	}
	val := s.send[s.idx]
	s.idx++
	return val, nil
}

func (s *testBidiServer) Send(kv *proto.KeyValue) error {
	s.receive = append(s.receive, kv)
	return nil
}

func (s *testBidiKeyServer) Send(kv *proto.KeyValue) error {
	s.receive = append(s.receive, kv)
	return nil
}

func oneThroughHundred() []*proto.KeyValue {
	lst := make([]*proto.KeyValue, 100)
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("%02d", i)
		lst[i] = &proto.KeyValue{Key: k, Value: []byte(k)}
	}
	return lst
}
