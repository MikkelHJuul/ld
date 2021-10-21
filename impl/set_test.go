package impl

import (
	"github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	"testing"
)

func Test_ldService_SetMany(t *testing.T) {
	tests := []struct {
		name   string
		server *testBidiServer
	}{
		{
			name: "a simple server",
			server: newTestServer([]*proto.KeyValue{
				{Key: []byte("1"), Value: []byte("1")},
				{Key: []byte("2"), Value: []byte("2")},
				{Key: []byte("3"), Value: []byte("3")},
				{Key: []byte("4"), Value: []byte("4")},
				{Key: []byte("5"), Value: []byte("5")},
			}),
		},
		{
			name: "a simple server re-set key",
			server: newTestServer([]*proto.KeyValue{
				{Key: []byte("1"), Value: []byte("1")},
				{Key: []byte("1"), Value: []byte("1")},
				{Key: []byte("1"), Value: []byte("4")},
				{Key: []byte("1"), Value: []byte("2")},
				{Key: []byte("1"), Value: []byte("1")},
				{Key: []byte("1"), Value: []byte("1")},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := NewServer(func(bo *badger.Options) {
				*bo = badger.DefaultOptions("").WithInMemory(true)
			})
			if err := l.SetMany(tt.server); err != nil {
				t.Errorf("SetMany() error = %v", err)
			}
			if len(tt.server.send) != len(tt.server.receive) {
				t.Errorf("mismatch in values received and sent")
			}
			for _, it := range tt.server.receive {
				if len(it.Key) > 0 && it.Value != nil {
					t.Errorf("non-nil return")
				}
			}
		})
	}
}
