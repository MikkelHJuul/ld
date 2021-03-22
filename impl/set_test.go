package impl

import (
	"github.com/MikkelHJuul/ld/proto"
	"testing"
)

func Test_ldService_SetMany(t *testing.T) {
	tests := []struct {
		name   string
		server *testBidiServer
	}{
		{
			name: "a simple server",
			server: NewTestServer([]*proto.KeyValue{
				{Key: "1", Value: []byte("1")},
				{Key: "2", Value: []byte("2")},
				{Key: "3", Value: []byte("3")},
				{Key: "4", Value: []byte("4")},
				{Key: "5", Value: []byte("5")},
			}),
		},
		{
			name: "a simple server re-set key",
			server: NewTestServer([]*proto.KeyValue{
				{Key: "1", Value: []byte("1")},
				{Key: "1", Value: []byte("1")},
				{Key: "1", Value: []byte("4")},
				{Key: "1", Value: []byte("2")},
				{Key: "1", Value: []byte("1")},
				{Key: "1", Value: []byte("1")},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewServer("", true)
			if err := l.SetMany(tt.server); err != nil {
				t.Errorf("SetMany() error = %v", err)
			}
			if len(tt.server.send) != len(tt.server.receive) {
				t.Errorf("mismatch in values received and sent")
			}
			for _, it := range tt.server.receive {
				if it != nil {
					t.Errorf("non-nil return")
				}
			}
		})
	}
}
