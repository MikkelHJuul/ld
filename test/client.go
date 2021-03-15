package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/MikkelHJuul/ld/test/client-proto"
	"github.com/mmcloughlin/geohash"
	"google.golang.org/grpc"
)

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

	stream, _ := client.SetMany(ctx)

	transferData := true
	go func() {
		if transferData {
			jsonFile, err := os.Open("2018-07.txt")
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
					Key:   keyFromMessage(&v),
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
		}
	}()
	for transferData {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil || !(msg == nil || msg.Value == nil) {
			log.Print("something went wrong")
		}
	}
	featureTime, _ := time.Parse(time.RFC3339, "2018-07-04T19:01:12.324000Z")
	timePrefix := make([]byte, 8)
	binary.LittleEndian.PutUint64(timePrefix, uint64(featureTime.Unix()))
	//keyPrefix := hex.EncodeToString(timePrefix)[:5] //remove a prefix value
	readStream, err := client.GetRange(ctx, &ld_proto.KeyRange{})
	if err != nil {
		log.Fatal(err)
	}
	features := make([]*ld_proto.KeyValue, 56000)
	for {
		feature, err := readStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
		}
		features = append(features, feature)
	}
	log.Printf("returned %d records, in %s seconds", len(features), time.Since(start))
}

func keyFromMessage(feature *ld_proto.Feature) string {
	geoHash := geohash.EncodeWithPrecision(feature.Geometry.Coordinates[1], feature.Geometry.Coordinates[0], 8)
	featureTime, _ := time.Parse(time.RFC3339, feature.Properties.Observed)
	timePrefix := make([]byte, 8)
	binary.LittleEndian.PutUint64(timePrefix, uint64(featureTime.Unix()))
	timeGeoHashPrefix := hex.EncodeToString(timePrefix)[:8] + geoHash
	amp := int(feature.Properties.Amp * 10)
	return fmt.Sprintf("%s%d%d", timeGeoHashPrefix, feature.Properties.Type, amp)
}
