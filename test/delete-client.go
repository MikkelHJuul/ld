package main

import (
	"context"
	pb "github.com/MikkelHJuul/ld/test/client-proto"
	"google.golang.org/grpc"
	"io"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:5326", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewLdClient(conn)
	ctx := context.Background()

	cl, err := client.DeleteRange(ctx, &pb.KeyRange{Pattern: "^00"})
	if err != nil {
		log.Printf("got meself an error: %v", err)
	}
	for {
		kv, err := cl.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("deleteStream broke! %v", err)
		}
		log.Println(kv)
	}
}