package impl

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	ldProto "github.com/MikkelHJuul/ld/proto"
	"github.com/desertbit/grumble"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
)

type deOrEncodeFunction func([]byte, *dynamic.Message) error
type getDynamicMessageDecode func([]byte, deOrEncodeFunction) (*dynamic.Message, error)
type executionMethod func(decode getDynamicMessageDecode) (*dynamic.Message, *ldProto.KeyValue, error)

func unmarshalJSON(bytes []byte, message *dynamic.Message) error {
	return message.UnmarshalJSON(bytes)
}

func unmarshal(bytes []byte, message *dynamic.Message) error {
	return message.Unmarshal(bytes)
}

func newClientAndCtx(ctx *grumble.Context, timeout time.Duration) (ldProto.LdClient, context.Context, func()) {
	conn, err := grpc.Dial(ctx.Flags.String("target"), grpc.WithInsecure())
	if err != nil {
		ctx.App.PrintError(errText("failed to dial server" + ctx.Flags.String("target")))
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

func getProtoMsgAndDecode(msg []byte, protofile string, m deOrEncodeFunction) (*dynamic.Message, error) {
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
		for _, msgDesc := range pf.GetMessageTypes() {

			dMsg = dynamic.NewMessage(msgDesc)
			if err = m(msg, dMsg); err == nil {
				break out
			}

		}
	}
	return dMsg, err
}

func exec(ctx *grumble.Context, cmd executionMethod) error {
	protoFile := ctx.Flags.String("protofile")
	dMsg, kv, err := cmd(func(b []byte, meth deOrEncodeFunction) (*dynamic.Message, error) {
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

func handleKeyValueReturned(val *ldProto.KeyValue) executionMethod {
	return func(decode getDynamicMessageDecode) (*dynamic.Message, *ldProto.KeyValue, error) {
		dMsg, err := decode(val.Value, unmarshal)
		if err != nil {
			return nil, val, errText("error creating dynamic message")
		}
		if dMsg != nil {
			msg, err := dMsg.MarshalJSON()
			if err != nil {
				return nil, nil, err
			}
			val.Value = msg
		}
		return nil, val, err
	}
}

// Get implements `ld.proto` service rpc `Get`
func Get(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, 5*time.Second)
	defer cancel()
	val, err := client.Get(execCtx, &ldProto.Key{Key: ctx.Args.String("key")})
	if err != nil || val.Key == "" {
		return err
	}
	return exec(ctx, handleKeyValueReturned(val))
}

// Set implements `ld.proto` service rpc `Set`
func Set(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, 5*time.Second)
	defer cancel()
	return exec(ctx, func(decode getDynamicMessageDecode) (*dynamic.Message, *ldProto.KeyValue, error) {
		msg := []byte(ctx.Args.String("value"))
		dMsg, err := decode(msg, unmarshalJSON)
		if err != nil {
			ctx.App.PrintError(errText("error creating dynamic message"))
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

// Delete implements `ld.proto` service rpc `Delete`
func Delete(ctx *grumble.Context) error {
	client, execCtx, cancel := newClientAndCtx(ctx, 5*time.Second)
	defer cancel()
	val, err := client.Delete(execCtx, &ldProto.Key{Key: ctx.Args.String("key")})
	if err != nil || val.Key == "" {
		return err
	}
	return exec(ctx, handleKeyValueReturned(val))
}

// GetRange implements `ld.proto` streaming service rpc `GetRange`
func GetRange(ctx *grumble.Context) error {
	return executeRange(ctx, "Received messages:", func(client ldProto.LdClient, execCtx context.Context, keyRange *ldProto.KeyRange) (rangeClient, error) {
		return client.GetRange(execCtx, keyRange)
	})
}

// DeleteRange implements `ld.proto` streaming service rpc `DeleteRange`
func DeleteRange(ctx *grumble.Context) error {
	return executeRange(ctx, "Deleted messages:", func(client ldProto.LdClient, execCtx context.Context, keyRange *ldProto.KeyRange) (rangeClient, error) {
		return client.DeleteRange(execCtx, keyRange)
	})
}

func executeRange(ctx *grumble.Context, msg string, ranger func(ldProto.LdClient, context.Context, *ldProto.KeyRange) (rangeClient, error)) error {
	client, execCtx, cancel := newClientAndCtx(ctx, time.Minute)
	defer cancel()
	keyRange := &ldProto.KeyRange{
		Prefix:  ctx.Flags.String("prefix"),
		Pattern: ctx.Flags.String("pattern"),
		From:    ctx.Flags.String("from"),
		To:      ctx.Flags.String("to"),
	}

	stream, err := ranger(client, execCtx, keyRange)
	if err == nil {
		var count int
		count, err = handleRangeStream(ctx, stream)
		ctx.App.Println(msg, count)
	}
	return err
}

type rangeClient interface {
	Recv() (*ldProto.KeyValue, error)
}

type errText string

func (X errText) Error() string {
	return string(X)
}

func handleRangeStream(ctx *grumble.Context, stream rangeClient) (int, error) {
	var dMsg *dynamic.Message
	protofile := ctx.Flags.String("protofile")
	count := 0
	msgChan := make(chan *ldProto.KeyValue)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	output := make(chan error, 1)

	go func() {
		for {
			kv := <-msgChan
			if dMsg == nil && protofile != "" {
				var err error
				dMsg, err = getProtoMsgAndDecode(kv.Value, protofile, func(b []byte, m *dynamic.Message) error {
					return m.Unmarshal(b)
				})
				if err != nil {
					ctx.App.PrintError(errText("error building dynamic message, cannot decode/serialize messages"))
				}
			}
			if dMsg != nil {
				dMsg.Reset()
				if err := dMsg.Unmarshal(kv.Value); err != nil {
					output <- err
					break
				}
				jsBytes, err := dMsg.MarshalJSON()
				if err != nil {
					output <- err
					break
				}
				kv.Value = jsBytes
			}
			_, _ = ctx.App.Println(kv)
			count++
		}
	}()

	var err error
	for {
		var keyValue *ldProto.KeyValue
		keyValue, err = stream.Recv()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}
		select {
		case <-interrupt:
			ctx.App.PrintError(errText("Interrupted by the user"))
			break
		case err = <-output:
			break
		case msgChan <- keyValue:
		}
	}
	return count, err
}
