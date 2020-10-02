package main

import (
	"github.com/MikkelHJuul/ld/data"
	pb "github.com/MikkelHJuul/ld/service"

	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
)

var (
	port = flag.String("port", lookupEnvOrString("PORT", "5326"), "The server port, default 5326")

	serviceType = flag.String("service-type", lookupEnvOrString("SERVICE_TYPE", "MMAP"), "the values MMAP or MEM, referring to mmap or completely in-memory storage")

	mmapFile = flag.String("mmap-file", lookupEnvOrString("MMAP_FILE", "/data/ld.dat"), "the path or file where data is stored, default '/data/ld.dat'")

	memSize = flag.String("mem-size", lookupEnvOrString("MEM_SIZE", "5G"), "The size of memory allowed to allocate (the data only)")
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:" + *port))
	if err != nil {
		_ = fmt.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterLdServer(grpcServer, newServer())
	if err := grpcServer.Serve(lis); err != nil {
		_ = fmt.Errorf("server exited with error: %v", err)
		os.Exit(1)
	}
}

type ldService struct {
	service data.Service
}

func newServer() *ldService {
	var service data.Service
	var mmapFileName = *mmapFile
	if *serviceType == "MEM" {
		service = data.NewCacheService(*memSize)
	} else {
		if info := fileExists(*mmapFile); info != nil {
			if info.IsDir() {
				mmapFileName += "/ld.dat"
			}
		} else {
			fmt.Println("file does not exist")
			if mmapFileName[len(mmapFileName)-1:] == "/" {
				mmapFileName += "ld.dat"
			}
		}
		service = data.NewMMapService(mmapFileName)
	}
	return &ldService{service: service}
}

func fileExists(filename string) os.FileInfo {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	}
	return info
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
		return stream.Send(keyValue)
	}
}

func (lds ldService) FetchMany(stream pb.Ld_FetchManyServer) error {
	return lds.handleMany(stream, lds.Fetch)
}

func (lds ldService) handleRange(method func(rng ...string) error, rng *pb.KeyRange) error {
	var ran []string
	if rng.Pattern != "" {
		ran = []string{rng.Pattern}
	} else {
		if rng.From != "" {
			ran = append(ran, rng.From)
		}
		if rng.To != "" {
			ran = append(ran, rng.To)
		}
	}

	return method(ran...)
}

func (lds ldService) FetchRange(rng *pb.KeyRange, stream pb.Ld_FetchRangeServer) error {
	methodToApply := func(rng ...string) error {
		return lds.service.GetRange(
			func(key string, bytes []byte) error {
				return stream.Send(&pb.KeyValue{Key: &pb.Key{Key: key}, Value: bytes})
			}, rng...)
	}
	return lds.handleRange(methodToApply, rng)
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
	methodToApply := func(rng ...string) error {
		return lds.service.DeleteRange(
			func(key string, bytes []byte) error {
				return stream.Send(&pb.KeyValue{Key: &pb.Key{Key: key}, Value: bytes})
			}, rng...)
	}
	return lds.handleRange(methodToApply, rng)
}

func (lds ldService) Insert(ctx context.Context, keyValue *pb.KeyValue) (*pb.InsertResponse, error) {
	err := lds.service.Save(keyValue.Key.Key, keyValue.Value)
	if err != nil {
		return &pb.InsertResponse{OK: false}, err
	}
	return &pb.InsertResponse{OK: true}, nil
}

func (lds ldService) InsertMany(stream pb.Ld_InsertManyServer) error {
	for {
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
