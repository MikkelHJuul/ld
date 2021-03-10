package main

import (
	"github.com/MikkelHJuul/ld/impl"
	pb "github.com/MikkelHJuul/ld/proto"

	"flag"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	port = flag.String("port", lookupEnvOrString("PORT", "5326"), "The server port, default 5326")

	mem = flag.Bool("in-mem", func() bool { _, ok := os.LookupEnv("IN_MEM"); return ok }(), "if the database is in-memory")
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:"+*port))
	if err != nil {
		_ = fmt.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	server := impl.NewServer(*mem)
	defer server.Close()
	pb.RegisterLdServer(grpcServer, server)
	if err := grpcServer.Serve(lis); err != nil {
		_ = fmt.Errorf("server exited with error: %v", err)
		os.Exit(1)
	}
}
