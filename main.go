package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/MikkelHJuul/ld/service"
	"google.golang.org/genproto/googleapis/ads/googleads/v4/errors"
	"google.golang.org/grpc"
	"io"
	"ldb/data"
	"log"
	"net"
	"os"
	"strconv"
)

var (
	prefix = flag.String("prefix", lookupEnvOrString("LD_PREFIX", ""), "use prefix for environment variables, remember '_'; default is empty string")
	port   = flag.Int("port", lookupEnvOrInt("PORT", 5326), "The server port, default 5326")

	storeType = flag.String("service-type", lookupEnvOrString("SERVICE_TYPE", "FS"), "the values FS or MEM, referring to file-based storage or in-memory storage")

	fsShardLevel = flag.Int("fs-shard-level", lookupEnvOrInt("FS_SHARD_LEVEL", 3), "The length of the sharding hierarchy, default 3")
	fsShardLen   = flag.Int("fs-shard-len", lookupEnvOrInt("FS_SHARD_LEN", 3), "how long the shard's length is, default 3")
	fsMemSize    = flag.Int("fs-mem-size", lookupEnvOrInt("FS_MEM_SIZE", 1000), "The number of cached items in the file-system type service's memory-cache, default 1000")
	fsRootPath   = flag.String("fs-root-path", lookupEnvOrString("FS_ROOT_PATH", "/data"), "the path where data is stored, default '/data'")

	memSize = flag.Int("mem-size", lookupEnvOrInt("MEM_SIZE", 100000), "The number of items to hold in memory, todo guide, default 100,000")
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(*prefix + key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(*prefix + key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterLdServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

type ldService struct {
	service data.Service
}

func newServer() *ldService { // pointers, which should I do??
	var service data.Service
	if *storeType == "MEM" {
		service = data.NewCacheService(*memSize)
	} else {
		service = data.NewFileService(*fsShardLen, *fsShardLevel, *fsRootPath, *fsMemSize)
	}
	return &ldService{service: service}
}

func (lds ldService) Fetch(ctx context.Context, key *pb.Key) (*pb.KeyValue, error) {
	val, err := lds.service.Get(key.Key)
	if val != nil {
		return &pb.KeyValue{
			Key:   key,
			Value: val,
		}, nil
	}
	return nil, err
}

func (lds ldService) handleMany(stream pb.Ld_FetchManyServer, method func(ctx context.Context, key *pb.Key) (*pb.KeyValue, error)) error {
	for {
		key, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		keyValue, err := method(nil, key)
		if err != nil {
			if err := stream.Send(nil); err != nil {
				return err
			}
		}
		if err := stream.Send(keyValue); err != nil {
			return err
		}
	}
}

func (lds ldService) FetchMany(stream pb.Ld_FetchManyServer) error {
	return lds.handleMany(stream, lds.Fetch)
}

func (lds ldService) FetchRange(rng *pb.KeyRange, stream pb.Ld_FetchRangeServer) error {
	return nil
}

func (lds ldService) Delete(ctx context.Context, key *pb.Key) (*pb.KeyValue, error) {
	val, err := lds.service.Delete(key.Key)
	if val != nil {
		return &pb.KeyValue{
			Key:   key,
			Value: val,
		}, nil
	}
	return nil, err
}

func (lds ldService) DeleteMany(stream pb.Ld_DeleteManyServer) error {
	return lds.handleMany(stream, lds.Delete)
}

func (lds ldService) DeleteRange(rng *pb.KeyRange, stream pb.Ld_DeleteRangeServer) error {
	var ran []string
	if rng.Pattern != "" {
		ran = []string{rng.Pattern}
	} else {
		if rng.From != "" {
			_ = append(ran, rng.From)
		}
		if rng.To != "" {
			_ = append(ran, rng.To)
		}
	}

	if err := lds.service.DeleteRange(
		func(key string, bytes []byte) error {
			if err := stream.Send(&pb.KeyValue{Key: &pb.Key{Key: key}, Value: bytes}); err != nil {
				return err
			}
			return nil
		},
		ran...); err != nil {
		return err
	}
}

func (lds ldService) Insert(ctx context.Context, key *pb.KeyValue) (*pb.InsertResponse, error) {
	err := lds.service.Save(key.Key.Key, key.Value)
	if err != nil {
		return &pb.InsertResponse{OK: false}, err
	}
	return &pb.InsertResponse{OK: true}, nil
}

func (lds ldService) InsertMany(stream pb.Ld_InsertManyServer) error {
	for { // duplicated-ish segment :(
		keyValue, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		iResp, err := lds.Insert(nil, keyValue)
		if err != nil {
			println(err)
			if err := stream.Send(iResp); err != nil {
				return err
			}
		}
		if err := stream.Send(iResp); err != nil {
			return err
		}
	}
}
