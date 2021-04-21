package impl

import (
	"github.com/MikkelHJuul/ld/proto"
	"testing"
)

func Test_ldService_GetMany(t *testing.T) {
	l := newTestBadger(t)
	tests := []struct {
		name    string
		server  *testBidiKeyServer
		results []*proto.KeyValue
	}{
		{
			name: "get some valid keys",
			server: newTestKeyServer([]*proto.Key{
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
			server: newTestKeyServer([]*proto.Key{
				{Key: "99"},
				{Key: "00"},
				{Key: "100"},
				{Key: "a"},
			}),
			results: []*proto.KeyValue{
				{Key: "99", Value: []byte("99")},
				{Key: "00", Value: []byte("00")},
				{},
				{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := l.GetMany(tt.server); err != nil {
				t.Errorf("GetMany() error = %v", err)
			}
			validateReturn(t, tt.results, tt.server.receive)
		})
	}
}

func Test_ldService_GetRange(t *testing.T) {
	l := newTestBadger(t)
	tests := []struct {
		name     string
		keyRange *proto.KeyRange
		response []*proto.KeyValue
	}{
		{
			name:     "get range within",
			keyRange: &proto.KeyRange{From: "12", To: "17"},
			response: []*proto.KeyValue{
				{Key: "12", Value: []byte("12")},
				{Key: "13", Value: []byte("13")},
				{Key: "14", Value: []byte("14")},
				{Key: "15", Value: []byte("15")},
				{Key: "16", Value: []byte("16")},
				{Key: "17", Value: []byte("17")},
			},
		},
		{
			name:     "get range overlap",
			keyRange: &proto.KeyRange{From: "99", To: "a"},
			response: []*proto.KeyValue{
				{Key: "99", Value: []byte("99")},
			},
		},
		{
			name:     "get prefix",
			keyRange: &proto.KeyRange{Prefix: "9"},
			response: []*proto.KeyValue{
				{Key: "90", Value: []byte("90")},
				{Key: "91", Value: []byte("91")},
				{Key: "92", Value: []byte("92")},
				{Key: "93", Value: []byte("93")},
				{Key: "94", Value: []byte("94")},
				{Key: "95", Value: []byte("95")},
				{Key: "96", Value: []byte("96")},
				{Key: "97", Value: []byte("97")},
				{Key: "98", Value: []byte("98")},
				{Key: "99", Value: []byte("99")},
			},
		},
		{
			name:     "get prefix From",
			keyRange: &proto.KeyRange{Prefix: "9", From: "92"},
			response: []*proto.KeyValue{
				{Key: "92", Value: []byte("92")},
				{Key: "93", Value: []byte("93")},
				{Key: "94", Value: []byte("94")},
				{Key: "95", Value: []byte("95")},
				{Key: "96", Value: []byte("96")},
				{Key: "97", Value: []byte("97")},
				{Key: "98", Value: []byte("98")},
				{Key: "99", Value: []byte("99")},
			},
		},
		{
			name:     "get prefix To",
			keyRange: &proto.KeyRange{Prefix: "9", To: "92"},
			response: []*proto.KeyValue{
				{Key: "90", Value: []byte("90")},
				{Key: "91", Value: []byte("91")},
				{Key: "92", Value: []byte("92")},
			},
		},
		{
			name:     "get prefix FromTo",
			keyRange: &proto.KeyRange{Prefix: "9", From: "91", To: "92"},
			response: []*proto.KeyValue{
				{Key: "91", Value: []byte("91")},
				{Key: "92", Value: []byte("92")},
			},
		},
		{
			name:     "get prefix Pattern",
			keyRange: &proto.KeyRange{Prefix: "9", Pattern: ".2"},
			response: []*proto.KeyValue{
				{Key: "92", Value: []byte("92")},
			},
		},
		{
			name:     "get Pattern",
			keyRange: &proto.KeyRange{Pattern: ".3"},
			response: []*proto.KeyValue{
				{Key: "03", Value: []byte("03")},
				{Key: "13", Value: []byte("13")},
				{Key: "23", Value: []byte("23")},
				{Key: "33", Value: []byte("33")},
				{Key: "43", Value: []byte("43")},
				{Key: "53", Value: []byte("53")},
				{Key: "63", Value: []byte("63")},
				{Key: "73", Value: []byte("73")},
				{Key: "83", Value: []byte("83")},
				{Key: "93", Value: []byte("93")},
			},
		},
		{
			name:     "get prefix and from to mismatch",
			keyRange: &proto.KeyRange{Prefix: "9", From: "12", To: "78"},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get prefix and from to mismatch",
			keyRange: &proto.KeyRange{Prefix: "5", From: "60"},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get prefix and from to mismatch",
			keyRange: &proto.KeyRange{Prefix: "5", To: "49"},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get range outside",
			keyRange: &proto.KeyRange{From: "a"},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get range to zero",
			keyRange: &proto.KeyRange{To: "0"},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get To one",
			keyRange: &proto.KeyRange{To: "1"},
			response: []*proto.KeyValue{
				{Key: "00", Value: []byte("00")},
				{Key: "01", Value: []byte("01")},
				{Key: "02", Value: []byte("02")},
				{Key: "03", Value: []byte("03")},
				{Key: "04", Value: []byte("04")},
				{Key: "05", Value: []byte("05")},
				{Key: "06", Value: []byte("06")},
				{Key: "07", Value: []byte("07")},
				{Key: "08", Value: []byte("08")},
				{Key: "09", Value: []byte("09")},
			},
		},
	}
	for _, tt := range tests {
		server := newTestServer(nil)
		t.Run(tt.name, func(t *testing.T) {
			if err := l.GetRange(tt.keyRange, server); err != nil {
				t.Errorf("GetRange() error = %v", err)
			}
			validateReturn(t, tt.response, server.receive)
		})
	}
}

func TestGetRangeReturnsError(t *testing.T) {
	l := newTestBadger(t)
	err := l.GetRange(&proto.KeyRange{Pattern: "("}, newTestServer(nil))
	if err == nil {
		t.Error("err shouldn't have been nil")
	}
}

func TestGetManyServerErrors(t *testing.T) {
	l := newTestBadger(t)
	err := l.GetMany(&erroringKeyServer{})
	if err == nil {
		t.Error("err shouldn't have been nil")
	}
}
