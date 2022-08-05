package main

import (
	"bytes"
	"context"
	"regexp"

	v1 "github.com/MikkelHJuul/ld/gen/ld/v1"
	connect_go "github.com/bufbuild/connect-go"
)

type sendServerStream[res any] struct {
	stream   *connect_go.ServerStream[res]
	check    func() error
	respMeth func(kv) *res
}

// Send implements sender
func (s *sendServerStream[res]) Send(keyVal kv) error {
	if err := s.check(); err != nil {
		return err
	}
	if err := s.stream.Send(s.respMeth(keyVal)); err != nil {
		return err
	}
	return nil
}

var _ sender = (*sendServerStream[any])(nil)

// DeleteRange implements ldv1connect.LdServiceHandler
func (lk *ldKv) DeleteRange(ctx context.Context, req *connect_go.Request[v1.DeleteRangeRequest], stream *connect_go.ServerStream[v1.DeleteRangeResponse]) error {
	rangeOp := rangerFromKeyRange(req.Msg.Range)
	streamer := &sendServerStream[v1.DeleteRangeResponse]{
		stream: stream,
		check:  checkCtx(ctx),
		respMeth: func(k kv) *v1.DeleteRangeResponse {
			return &v1.DeleteRangeResponse{Kv: &v1.KeyValue{Key: k.key, Value: k.value}}
		},
	}
	return lk.db.deleteRange(rangeOp, streamer)
}

// GetRange implements ldv1connect.LdServiceHandler
func (lk *ldKv) GetRange(ctx context.Context, req *connect_go.Request[v1.GetRangeRequest], stream *connect_go.ServerStream[v1.GetRangeResponse]) error {
	rangeOp := rangerFromKeyRange(req.Msg.Range)
	streamer := &sendServerStream[v1.GetRangeResponse]{
		stream: stream,
		check:  checkCtx(ctx),
		respMeth: func(k kv) *v1.GetRangeResponse {
			return &v1.GetRangeResponse{Kv: &v1.KeyValue{Key: k.key, Value: k.value}}
		},
	}
	return lk.db.getRange(rangeOp, streamer)
}

type rangerFeatures int

const none rangerFeatures = 0
const (
	lower rangerFeatures = 1 << iota
	upper
	prefix
	backward
	//forward is unused because it's default
)

var feats [16]rangerGen

func init() {
	opts := func(r ...RangerOpt) []RangerOpt {
		return r
	}

	var rangers = map[rangerFeatures]rangerGen{
		none: func(kr *v1.KeyRange) []RangerOpt {
			return opts()
		},
		lower: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithSeek(SeekerTo(kr.Lower)))
		},
		upper: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithValidatorMax(kr.Upper))
		},
		prefix: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithPrefix(kr.Prefix))
		},
		backward: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithScanner(ReverseScanning()))
		},
		lower ^ upper: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithSeek(SeekerTo(kr.Lower)), WithValidatorMax(kr.Upper))
		},
		lower ^ backward: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithSeek(SeekerLast()), WithValidatorMin(kr.Lower), WithScanner(ReverseScanning()))
		},
		lower ^ prefix: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithSeek(SeekerTo(kr.Lower)), WithValidatorPrefix(kr.Prefix))
		},
		lower ^ upper ^ backward: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithSeek(SeekerTo(kr.Upper)), WithValidatorMin(kr.Lower), WithScanner(ReverseScanning()))
		},
		lower ^ prefix ^ backward: func(kr *v1.KeyRange) []RangerOpt {
			if bytes.Compare(kr.Lower, kr.Prefix) > 0 {
				return opts(WithSeek(SeekerTo(plusOne(kr.Prefix))), WithValidatorMin(kr.Lower), WithScanner(ReverseScanning()))
			}
			return opts(WithSeek(SeekerTo(plusOne(kr.Prefix))), WithValidatorPrefix(kr.Prefix), WithScanner(ReverseScanning()))
		},
		lower ^ upper ^ prefix: func(kr *v1.KeyRange) []RangerOpt {
			var rOpts []RangerOpt
			if bytes.Compare(kr.Upper, kr.Prefix) > 0 {
				rOpts = append(rOpts, WithValidatorPrefix(kr.Prefix))
			} else {
				rOpts = append(rOpts, WithValidatorMax(kr.Upper))
			}
			if bytes.Compare(kr.Lower, kr.Prefix) > 0 {
				rOpts = append(rOpts, WithSeek(SeekerTo(kr.Lower)))
			} else {
				rOpts = append(rOpts, WithSeek(SeekerTo(kr.Prefix)))
			}
			return opts(rOpts...)
		},
		lower ^ upper ^ prefix ^ backward: func(kr *v1.KeyRange) []RangerOpt {
			var rOpts []RangerOpt
			if bytes.Compare(kr.Upper, kr.Prefix) > 0 {
				rOpts = append(rOpts, WithSeek(SeekerTo(kr.Prefix)))
			} else {
				rOpts = append(rOpts, WithSeek(SeekerTo(kr.Upper)))
			}
			if bytes.Compare(kr.Lower, kr.Prefix) > 0 {
				rOpts = append(rOpts, WithValidatorMin(kr.Lower))
			} else {
				rOpts = append(rOpts, WithValidatorPrefix(kr.Prefix))
			}
			return opts(rOpts...)
		},
		upper ^ prefix: func(kr *v1.KeyRange) []RangerOpt {
			if bytes.Compare(kr.Upper, kr.Prefix) > 0 {
				return opts(WithPrefix(kr.Prefix))
			}
			return opts(WithSeek(SeekerTo(kr.Prefix)), WithValidatorMax(kr.Upper))
		},
		upper ^ backward: func(kr *v1.KeyRange) []RangerOpt {
			return opts(WithSeek(SeekerTo(kr.Upper)), WithScanner(ReverseScanning()))
		},
		upper ^ prefix ^ backward: func(kr *v1.KeyRange) []RangerOpt {
			if bytes.Compare(kr.Upper, kr.Prefix) > 0 {
				return opts(WithSeek(SeekerTo(plusOne(kr.Prefix))), WithValidatorPrefix(kr.Prefix), WithScanner(ReverseScanning()))
			}
			return opts(WithSeek(SeekerTo(kr.Upper)), WithValidatorPrefix(kr.Prefix), WithScanner(ReverseScanning()))
		},
		prefix ^ backward: func(kr *v1.KeyRange) []RangerOpt { //fix Seekto off by one, first element
			return opts(WithSeek(SeekerTo(plusOne(kr.Prefix))), WithValidatorPrefix(kr.Prefix), WithScanner(ReverseScanning()))
		},
	}

	for k, v := range rangers {
		feats[k] = v
	}
}

func plusOne(b []byte) []byte {
	cop := make([]byte, len(b))
	_ = copy(cop, b)
	cop[len(b)] = uint8(cop[len(b)]) + 1
	return cop
}

type rangerGen func(*v1.KeyRange) []RangerOpt

func rangerFromKeyRange(keyRange *v1.KeyRange) ranger {
	var features rangerFeatures = featuresFrom(keyRange)
	r := feats[features](keyRange)
	if keyRange.Pattern != "" {
		reg, _ := regexp.Compile(keyRange.Pattern) // TODO err handling
		r = append(r, WithAcceptor(func(k key) bool {
			return reg.Match(k)
		}))
	}
	return NewRanger(r...)
}

func featuresFrom(keyRange *v1.KeyRange) rangerFeatures {
	feats := none
	if keyRange.Lower != nil {
		feats ^= lower
	}
	if keyRange.Upper != nil {
		feats ^= upper
	}
	if keyRange.Prefix != nil {
		feats ^= prefix
	}
	if keyRange.Direction == v1.KeyRange_REVERSE {
		feats ^= backward
	}
	return feats
}
