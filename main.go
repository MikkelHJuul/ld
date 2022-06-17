package main

import (
	"errors"
	"flag"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

var (
	port       = flag.String("port", lookupEnvOrString("PORT", "5326"), "The server port, default 5326")
	dbLocation = flag.String("db-location", lookupEnvOrString("DB_LOCATION", "ld_db"), "folder location where the database is situated")
	logLevel   = flag.String("log-level", lookupEnvOrString("LOG_LEVEL", "INFO"), "configure logging level")
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
	_, err = os.Stat(*dbLocation)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf(`could not open file, does not exist:`, *dbLocation)
		} else {
			log.Fatalf(`unexpected error with database location:`, *dbLocation)
		}
	}
	var opts []grpc.ServerOption
	var l = ldServer{make(map[string]*service)}
	opts = append(opts, grpc.UnknownServiceHandler(l.serveUnknownService()))
	// create general server
	// use grpc.UnkownServiceHandler to service requests (it's hacky! but it works)
	// link this and admin server somehow (how tight should the bound be?)

	grpcServer := grpc.NewServer(opts...) // unkown server is injected via options

	// create and register admin server

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
