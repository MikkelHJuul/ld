package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/MikkelHJuul/ld/test/client-proto"
	"google.golang.org/grpc"
)

// This client loads a Line-delimited json-file of format ./client-proto/my_message.proto.
// It loads it into the database using a SetMany RPC and queries via Pattern.
// The json file can be downloaded via the DMI portal for free data (requires sign up)
// at https://dmigw.govcloud.dk/v2/lightningdata/bulk/?api-key=<your-key>
func main() {
	start := time.Now()
	conn, err := grpc.Dial("localhost:5326", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := ld_proto.NewLdClient(conn)
	ctx, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()

	//timeString := ld_proto.HandmadeTimeKeyString("2018-07-04T19:01:12.324000Z")
	readStream, err := client.GetRange(ctx, &ld_proto.KeyRange{Prefix: "00511", Pattern: "00511....u2"})
	if err != nil {
		log.Fatal(err)
	}
	//features := make([]*ld_proto.KeyValue, 0)
	count := 0
	for {
		feature, err := readStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		//features = append(features, feature)
		log.Print(feature)
		count++
	}
	log.Printf("returned %d records, in %s seconds", count, time.Since(start))
}
