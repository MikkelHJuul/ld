package impl

import (
	"bytes"
	"fmt"
	"github.com/MikkelHJuul/ld/proto"
	"google.golang.org/grpc"
	"io"
	"testing"
)

type testBidiServer struct {
	grpc.ServerStream
	send    []*proto.KeyValue
	receive []*proto.KeyValue
	idx     int
}

type testBidiKeyServer struct {
	grpc.ServerStream
	send    []*proto.Key
	receive []*proto.KeyValue
	idx     int
}

func newTestKeyServer(ks []*proto.Key) *testBidiKeyServer {
	return &testBidiKeyServer{
		send:    ks,
		receive: make([]*proto.KeyValue, 0),
		idx:     0,
	}
}

func newTestServer(kvs []*proto.KeyValue) *testBidiServer {
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

func newTestBadger(t *testing.T) *ldService {
	l := NewServer("", true)
	err := l.SetMany(newTestServer(oneThroughHundred()))
	if err != nil {
		t.Error("could not initiate test database")
	}
	return l
}

func validateReturn(t *testing.T, expected, got []*proto.KeyValue) {
	if len(got) != len(expected) {
		t.Errorf("not the same amount of results, %d =|= %d", len(expected), len(got))
	}
	numNils := 0
	for _, aVal := range got {
		if aVal == nil {
			numNils++
			continue
		}
		isThere := false
		for _, res := range expected {
			if aVal.Key == res.Key && bytes.Equal(aVal.Value, res.Value) {
				isThere = true
				break
			}
		}
		if !isThere {
			t.Errorf("results are not like! %v != %v", got, expected)
		}
	}
	for _, res := range expected {
		if res == nil {
			numNils--
		}
	}
	if numNils != 0 {
		t.Errorf("incorrect numbers of empty messages as expected")
	}
}
