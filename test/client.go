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
	conn, err := grpc.Dial("localhost:5326", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := ld_proto.NewLdClient(conn)
	ctx, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()

	if false {
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
			resp, err := client.Create(ctx, &ld_proto.KeyValue{
				Key:   keyFromMessage(&v),
				Value: &v,
			})
			if err != nil {
				log.Printf("send error")
			}
			if resp != nil && resp.Error {
				log.Printf("server error")
			}
		}
		if s.Err() != nil {
			log.Fatalf("some error %v", s.Err())
		}
	}
	featureTime, _ := time.Parse(time.RFC3339, "2018-07-04T19:01:12.324000Z")
	timePrefix := make([]byte, 8)
	binary.LittleEndian.PutUint64(timePrefix, uint64(featureTime.Unix()))
	keyPrefix := hex.EncodeToString(timePrefix)[:6] //remove a prefix value
	stream, err := client.ReadRange(ctx, &ld_proto.KeyRange{
		Prefix: keyPrefix,
	})
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
		}
		log.Printf("%s", feature.Key)
		count++
	}
	log.Printf("returned %d records", count)
}

func keyFromMessage(feature *ld_proto.Feature) string {
	geoHash := geohash.EncodeWithPrecision(feature.Geometry.Coordinates[1], feature.Geometry.Coordinates[0], 8)
	featureTime, _ := time.Parse(time.RFC3339, feature.Properties.Observed)
	timePrefix := make([]byte, 8)
	binary.LittleEndian.PutUint64(timePrefix, uint64(featureTime.Unix()))
	timeGeoHashPrefix := hex.EncodeToString(timePrefix) + geoHash
	return fmt.Sprintf("%s%d%f", timeGeoHashPrefix, feature.Properties.Type, feature.Properties.Amp)
}
