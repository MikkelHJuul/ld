package main

import (
	"flag"
	"github.com/MikkelHJuul/ld/impl"
	"github.com/MikkelHJuul/ld/proto"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"net"
	"os"
)

var (
	port       = flag.String("port", lookupEnvOrString("PORT", "5326"), "The server port, default 5326")
	dbLocation = flag.String("db-location", lookupEnvOrString("DB_LOCATION", "ld_badger"), "folder location where the database is situated")
	logLevel   = flag.String("log-level", lookupEnvOrString("LOG_LEVEL", "INFO"), "configure logging level")
	mem        = flag.Bool("in-mem", func() bool { _, ok := os.LookupEnv("IN_MEM"); return ok }(), "if the database is in-memory")
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	flag.Parse()
	loggingLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(loggingLevel)
	lis, err := net.Listen("tcp", "localhost:"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	if *mem {
		log.Infof("ignoring db-location: %s, because instance is set to run in-memory", dbLocation)
		*dbLocation = ""
	}
	server := impl.NewServer(
		func(bo *badger.Options) {
			*bo = badger.DefaultOptions(*dbLocation).WithInMemory(*mem)
		})
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
