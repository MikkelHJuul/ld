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
				{Key: []byte("04")},
				{Key: []byte("40")},
				{Key: []byte("99")},
				{Key: []byte("00")},
				{Key: []byte("22")},
			}),
			results: []*proto.KeyValue{
				{Key: []byte("04"), Value: []byte("04")},
				{Key: []byte("40"), Value: []byte("40")},
				{Key: []byte("99"), Value: []byte("99")},
				{Key: []byte("00"), Value: []byte("00")},
				{Key: []byte("22"), Value: []byte("22")},
			},
		},
		{
			name: "get invalid keys",
			server: newTestKeyServer([]*proto.Key{
				{Key: []byte("99")},
				{Key: []byte("00")},
				{Key: []byte("100")},
				{Key: []byte("a")},
			}),
			results: []*proto.KeyValue{
				{Key: []byte("99"), Value: []byte("99")},
				{Key: []byte("00"), Value: []byte("00")},
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
			keyRange: &proto.KeyRange{From: []byte("12"), To: []byte("17")},
			response: []*proto.KeyValue{
				{Key: []byte("12"), Value: []byte("12")},
				{Key: []byte("13"), Value: []byte("13")},
				{Key: []byte("14"), Value: []byte("14")},
				{Key: []byte("15"), Value: []byte("15")},
				{Key: []byte("16"), Value: []byte("16")},
				{Key: []byte("17"), Value: []byte("17")},
			},
		},
		{
			name:     "get range overlap",
			keyRange: &proto.KeyRange{From: []byte("99"), To: []byte("a")},
			response: []*proto.KeyValue{
				{Key: []byte("99"), Value: []byte("99")},
			},
		},
		{
			name:     "get prefix",
			keyRange: &proto.KeyRange{Prefix: []byte("9")},
			response: []*proto.KeyValue{
				{Key: []byte("90"), Value: []byte("90")},
				{Key: []byte("91"), Value: []byte("91")},
				{Key: []byte("92"), Value: []byte("92")},
				{Key: []byte("93"), Value: []byte("93")},
				{Key: []byte("94"), Value: []byte("94")},
				{Key: []byte("95"), Value: []byte("95")},
				{Key: []byte("96"), Value: []byte("96")},
				{Key: []byte("97"), Value: []byte("97")},
				{Key: []byte("98"), Value: []byte("98")},
				{Key: []byte("99"), Value: []byte("99")},
			},
		},
		{
			name:     "get prefix From",
			keyRange: &proto.KeyRange{Prefix: []byte("9"), From: []byte("92")},
			response: []*proto.KeyValue{
				{Key: []byte("92"), Value: []byte("92")},
				{Key: []byte("93"), Value: []byte("93")},
				{Key: []byte("94"), Value: []byte("94")},
				{Key: []byte("95"), Value: []byte("95")},
				{Key: []byte("96"), Value: []byte("96")},
				{Key: []byte("97"), Value: []byte("97")},
				{Key: []byte("98"), Value: []byte("98")},
				{Key: []byte("99"), Value: []byte("99")},
			},
		},
		{
			name:     "get prefix To",
			keyRange: &proto.KeyRange{Prefix: []byte("9"), To: []byte("92")},
			response: []*proto.KeyValue{
				{Key: []byte("90"), Value: []byte("90")},
				{Key: []byte("91"), Value: []byte("91")},
				{Key: []byte("92"), Value: []byte("92")},
			},
		},
		{
			name:     "get prefix FromTo",
			keyRange: &proto.KeyRange{Prefix: []byte("9"), From: []byte("91"), To: []byte("92")},
			response: []*proto.KeyValue{
				{Key: []byte("91"), Value: []byte("91")},
				{Key: []byte("92"), Value: []byte("92")},
			},
		},
		{
			name:     "get prefix Pattern",
			keyRange: &proto.KeyRange{Prefix: []byte("9"), Pattern: ".2"},
			response: []*proto.KeyValue{
				{Key: []byte("92"), Value: []byte("92")},
			},
		},
		{
			name:     "get Pattern",
			keyRange: &proto.KeyRange{Pattern: ".3"},
			response: []*proto.KeyValue{
				{Key: []byte("03"), Value: []byte("03")},
				{Key: []byte("13"), Value: []byte("13")},
				{Key: []byte("23"), Value: []byte("23")},
				{Key: []byte("33"), Value: []byte("33")},
				{Key: []byte("43"), Value: []byte("43")},
				{Key: []byte("53"), Value: []byte("53")},
				{Key: []byte("63"), Value: []byte("63")},
				{Key: []byte("73"), Value: []byte("73")},
				{Key: []byte("83"), Value: []byte("83")},
				{Key: []byte("93"), Value: []byte("93")},
			},
		},
		{
			name:     "get prefix and from to mismatch",
			keyRange: &proto.KeyRange{Prefix: []byte("9"), From: []byte("12"), To: []byte("78")},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get prefix and from to mismatch",
			keyRange: &proto.KeyRange{Prefix: []byte("5"), From: []byte("60")},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get prefix and from to mismatch",
			keyRange: &proto.KeyRange{Prefix: []byte("5"), To: []byte("49")},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get range outside",
			keyRange: &proto.KeyRange{From: []byte("a")},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get range to zero",
			keyRange: &proto.KeyRange{To: []byte("0")},
			response: []*proto.KeyValue{},
		},
		{
			name:     "get To one",
			keyRange: &proto.KeyRange{To: []byte("1")},
			response: []*proto.KeyValue{
				{Key: []byte("00"), Value: []byte("00")},
				{Key: []byte("01"), Value: []byte("01")},
				{Key: []byte("02"), Value: []byte("02")},
				{Key: []byte("03"), Value: []byte("03")},
				{Key: []byte("04"), Value: []byte("04")},
				{Key: []byte("05"), Value: []byte("05")},
				{Key: []byte("06"), Value: []byte("06")},
				{Key: []byte("07"), Value: []byte("07")},
				{Key: []byte("08"), Value: []byte("08")},
				{Key: []byte("09"), Value: []byte("09")},
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
