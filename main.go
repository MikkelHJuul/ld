package main

import (
	"github.com/MikkelHJuul/ld/impl"
	"github.com/MikkelHJuul/ld/proto"
	log "github.com/sirupsen/logrus"

	"flag"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	port     = flag.String("port", lookupEnvOrString("PORT", "5326"), "The server port, default 5326")
	logLevel = flag.String("log-level", lookupEnvOrString("LOG_LEVEL", "INFO"), "configure logging level")
	mem      = flag.Bool("in-mem", func() bool { _, ok := os.LookupEnv("IN_MEM"); return ok }(), "if the database is in-memory")
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	loggingLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(loggingLevel)
	flag.Parse()
	lis, err := net.Listen("tcp", "localhost:"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	server := impl.NewServer("ld_badger", *mem)
	defer func() {
		if err := server.Close(); err != nil {
			log.Error("error when closing the database", err)
		}
	}()
	proto.RegisterLdServer(grpcServer, server)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
