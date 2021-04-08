package impl

import (
	"context"
	ldProto "github.com/MikkelHJuul/ld/proto"
	"github.com/desertbit/grumble"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"io"
	"time"
)

func newClientAndCtx(ctx *grumble.Context, timeout time.Duration) (ldProto.LdClient, context.Context, func()) {
	conn, err := grpc.Dial(ctx.Flags.String("target"), grpc.WithInsecure())
	if err != nil {
		_, _ = ctx.App.Println("failed to dial server" + ctx.Flags.String("target"))
		return nil, nil, nil
	}

	client := ldProto.NewLdClient(conn)
	execCtx, cancel := context.WithTimeout(context.TODO(), timeout)
	cncl := func() {
		cancel()
		conn.Close()
	}
	return client, execCtx, cncl
}

func getProtoMsgAndDecode(msg []byte, protofile string, m func([]byte, *dynamic.Message) error) (*dynamic.Message, error) {
	if protofile == "" {
		return nil, nil
	}

	fds, err := protoparse.Parser{}.ParseFiles(protofile)
	if err != nil {
		return nil, err
	}
	var dMsg *dynamic.Message
out:
	for _, pf := range fds {
		for _, msgs := range pf.GetMessageTypes() {

			dMsg = dynamic.NewMessage(msgs)
			if err = m(msg, dMsg); err == nil {
				break out
			}

		}
	}
	return dMsg, err
}

func exec(ctx *grumble.Context, cmd func(func([]byte, func([]byte, *dynamic.Message) error) (*dynamic.Message, error)) (*dynamic.Message, *ldProto.KeyValue, error)) error {
	protoFile := ctx.Flags.String("protofile")
	dMsg, kv, err := cmd(func(b []byte, meth func([]byte, *dynamic.Message) error) (*dynamic.Message, error) {
		return getProtoMsgAndDecode(b, protoFile, meth)
	})
	if err != nil || kv == nil {
		return err
	}
	if dMsg != nil {
		//map value using the protofile
		if err = dMsg.Unmarshal(kv.Value); err != nil {
			return err
		}
		if txt, err := dMsg.MarshalJSON(); err == nil {
			kv = &ldProto.KeyValue{Key: kv.Key, Value: txt}
		} else {
			return err
		}
	}
	_, _ = ctx.App.Println(kv)
	return err
}

func Get(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, 5*time.Second)
	defer cancel()
	return exec(ctx, func(dynFun func([]byte, func([]byte, *dynamic.Message) error) (*dynamic.Message, error)) (*dynamic.Message, *ldProto.KeyValue, error) {
		key := ctx.Args.String("key")
		val, err := client.Get(execCtx, &ldProto.Key{Key: key})
		if err != nil || val.Key == "" {
			return nil, nil, err
		}
		dMsg, err := dynFun(val.Value, func(bytes []byte, message *dynamic.Message) error {
			return message.Unmarshal(bytes)
		})
		if err != nil {
			ctx.App.Println("error creating dynamic message")
			return nil, val, err
		}
		if dMsg != nil {
			msg, err := dMsg.MarshalJSON()
			if err != nil {
				return nil, nil, err
			}
			val.Value = msg
		}
		return nil, val, err
	})
}

func Set(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, 5*time.Second)
	defer cancel()
	return exec(ctx, func(dynFun func([]byte, func([]byte, *dynamic.Message) error) (*dynamic.Message, error)) (*dynamic.Message, *ldProto.KeyValue, error) {
		msg := []byte(ctx.Args.String("value"))
		dMsg, err := dynFun(msg, func(bytes []byte, message *dynamic.Message) error {
			return message.UnmarshalJSON(bytes)
		})
		if err != nil {
			ctx.App.Println("error creating dynamic message")
			return nil, nil, err
		}
		if dMsg != nil {
			msg, err = dMsg.Marshal()
			if err != nil {
				return nil, nil, err
			}
		}
		retMsg, err := client.Set(execCtx,
			&ldProto.KeyValue{
				Key:   ctx.Args.String("key"),
				Value: msg,
			})
		return dMsg, retMsg, err
	})
}

func Delete(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, 5*time.Second)
	defer cancel()
	return exec(ctx, func(dynFun func([]byte, func([]byte, *dynamic.Message) error) (*dynamic.Message, error)) (*dynamic.Message, *ldProto.KeyValue, error) {
		val, err := client.Delete(execCtx, &ldProto.Key{Key: ctx.Args.String("key")})
		if err != nil || val.Key == "" {
			return nil, nil, err
		}
		dMsg, err := dynFun(val.Value, func(bytes []byte, message *dynamic.Message) error {
			return message.Unmarshal(bytes)
		})
		if err != nil {
			ctx.App.Println("error creating dynamic message")
			return nil, val, err
		}
		if dMsg != nil {
			msg, err := dMsg.MarshalJSON()
			if err != nil {
				return nil, nil, err
			}
			val.Value = msg
		}
		return nil, val, err
	})
}

func GetRange(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, time.Minute)
	defer cancel()
	keyRange := &ldProto.KeyRange{
		Prefix:  ctx.Flags.String("prefix"),
		Pattern: ctx.Flags.String("pattern"),
		From:    ctx.Flags.String("from"),
		To:      ctx.Flags.String("to"),
	}

	if stream, err := client.GetRange(execCtx, keyRange); err != nil {
		return err
	} else {
		count, err := handleRangeStream(ctx, stream)
		if err != nil {
			return err
		}
		ctx.App.Println("received messages:", count)
	}
	return nil
}

func DeleteRange(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, time.Minute)
	defer cancel()
	keyRange := &ldProto.KeyRange{
		Prefix:  ctx.Flags.String("prefix"),
		Pattern: ctx.Flags.String("pattern"),
		From:    ctx.Flags.String("from"),
		To:      ctx.Flags.String("to"),
	}

	if stream, err := client.DeleteRange(execCtx, keyRange); err != nil {
		return err
	} else {
		count, err := handleRangeStream(ctx, stream)
		if err != nil {
			return err
		}
		ctx.App.Println("Deleted messages:", count)
	}
	return nil
}

type rangeClient interface {
	Recv() (*ldProto.KeyValue, error)
}

func handleRangeStream(ctx *grumble.Context, stream rangeClient) (int, error) {
	var dMsg *dynamic.Message
	protofile := ctx.Flags.String("protofile")
	count := 0
	for {
		kv, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		if dMsg == nil && protofile != "" {
			dMsg, err = getProtoMsgAndDecode(kv.Value, protofile, func(b []byte, m *dynamic.Message) error {
				return m.Unmarshal(b)
			})
		}
		if dMsg != nil {
			dMsg.Reset()
			if err := dMsg.Unmarshal(kv.Value); err != nil {
				return 0, err
			}
			jsBytes, err := dMsg.MarshalJSON()
			if err != nil {
				return 0, err
			}
			kv.Value = jsBytes
		}
		_, _ = ctx.App.Println(kv)
		count++
	}
	return count, nil
}
