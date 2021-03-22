package impl

import (
	"bytes"
	"github.com/MikkelHJuul/ld/proto"
	"testing"
)

func Test_ldService_GetMany(t *testing.T) {
	l := NewServer("", true)
	err := l.SetMany(NewTestServer(oneThroughHundred()))
	if err != nil {
		t.Errorf("could not initiate test database")
	}
	tests := []struct {
		name    string
		server  *testBidiKeyServer
		results []*proto.KeyValue
	}{
		{
			name: "get some valid keys",
			server: NewTestKeyServer([]*proto.Key{
				{Key: "04"},
				{Key: "40"},
				{Key: "99"},
				{Key: "00"},
				{Key: "22"},
			}),
			results: []*proto.KeyValue{
				{Key: "04", Value: []byte("04")},
				{Key: "40", Value: []byte("40")},
				{Key: "99", Value: []byte("99")},
				{Key: "00", Value: []byte("00")},
				{Key: "22", Value: []byte("22")},
			},
		},
		{
			name: "get invalid keys",
			server: NewTestKeyServer([]*proto.Key{
				{Key: "99"},
				{Key: "00"},
				{Key: "100"},
				{Key: "a"},
			}),
			results: []*proto.KeyValue{
				{Key: "99", Value: []byte("99")},
				{Key: "00", Value: []byte("00")},
				nil,
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := l.GetMany(tt.server); err != nil {
				t.Errorf("GetMany() error = %v", err)
			}
			if len(tt.server.receive) != len(tt.results) {
				t.Errorf("not the same amount of results, %d =|= %d", len(tt.server.receive), len(tt.results))
			}
			numNils := 0
			for _, aVal := range tt.server.receive {
				if aVal == nil {
					numNils++
					continue
				}
				isThere := false
				for _, res := range tt.results {
					if aVal.Key == res.Key && bytes.Equal(aVal.Value, res.Value) {
						isThere = true
						break
					}
				}
				if !isThere {
					t.Errorf("results are not like! %v != %v", tt.server.receive, tt.results)
				}
			}
			for _, res := range tt.results {
				if res == nil {
					numNils--
				}
			}
			if numNils != 0 {
				t.Errorf("incorrect numbers of empty messages as expected")
			}
		})
	}
}

func Test_ldService_GetRange(t *testing.T) {
	l := NewServer("", true)
	err := l.SetMany(NewTestServer(oneThroughHundred()))
	if err != nil {
		t.Errorf("could not initiate test database")
	}
	tests := []struct {
		name     string
		server   *testBidiServer
		keyRange *proto.KeyRange
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := l.GetRange(tt.keyRange, tt.server); err != nil {
				t.Errorf("GetRange() error = %v", err)
			}

		})
	}
}
