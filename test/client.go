package main

import (
	"context"
	"fmt"
	pb "github.com/MikkelHJuul/ld/test/client-proto"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"time"
)

var s1 = rand.NewSource(time.Now().UnixNano())
var r1 = rand.New(s1)

func randomDateInDecade() string {
	year := r1.Intn(9)
	mth := r1.Intn(11) + 1
	day := r1.Intn(29) + 1
	return fmt.Sprintf("%02d%02d%02d", year, mth, day)
}

func main() {
	conn, err := grpc.Dial("localhost:5326", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewLdClient(conn)
	ctx := context.Background()
	for i  := 0; i < 1000; i++ {
		for j  := 0; j < 200; j++ {

			key := randomDateInDecade() + fmt.Sprintf("%04d%03d", i, j) + "john"
			if _, err := client.Insert(ctx, &pb.KeyValue{
				Key: &pb.Key{Key: key},
				Value: &pb.YourObject{
					John: &pb.John{
						Name:       "John",
						Occupation: "being john",
					},
					JohnsApprentice: &pb.John{
						Name:       "not even closely John",
						Occupation: "being everyone else than john",
					},
				},
			}); err != nil {
				log.Fatalf("couldn't insert john: %v", err)
			} else {
				//log.Printf("Got a response, gee I hope it's OK: %v", *resp)
			}

			if _, err := client.Fetch(ctx, &pb.Key{Key: key}); err != nil {
				log.Fatalf("couldn't fetch john: %v", err)
			} else {
				//log.Printf("Got a response, gee I hope it's john: %v", *resp)
			}
		}
	}

}
