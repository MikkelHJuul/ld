package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/MikkelHJuul/ld/test/client-proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()
	conn, err := grpc.Dial("localhost:5326", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := ld_proto.NewLdClient(conn)
	ctx := context.TODO()

	stream, _ := client.SetMany(ctx)

	go func() {
		jsonFile, err := os.Open("all.txt")
		if err != nil {
			log.Fatalf("could not open file!")
		}

		fmt.Println("Successfully Opened json")
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		s := bufio.NewScanner(jsonFile)
		for s.Scan() {
			var v ld_proto.Feature
			if err := json.Unmarshal(s.Bytes(), &v); err != nil {
				log.Printf("unmarshal error")
			}
			err := stream.Send(&ld_proto.KeyValue{
				Key:   ld_proto.KeyFromMessage(&v),
				Value: &v,
			})
			if err != nil {
				log.Printf("send error")
			}
		}
		if s.Err() != nil {
			log.Fatalf("some error %v", s.Err())
		}
		_ = stream.CloseSend()
	}()
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil || !(msg == nil || msg.Value == nil) {
			log.Print("something went wrong")
		}
	}
	log.Printf("done writing to ld, duration: %s", time.Since(start))

}
