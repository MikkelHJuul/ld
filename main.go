package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"

	pb "github.com/MikkelHJuul/ld/service"
)

var (
	port = flag.Int("port", 5326, "The server port")
)

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

func newServer() pb.LdServer {
	return nil
}
