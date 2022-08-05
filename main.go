package main

import (
	"flag"
	"net/http"
	"os"
)

var (
	port       = flag.String("port", lookupEnvOrString("PORT", "5326"), "The server port, default 5326")
	dbLocation = flag.String("db-location", lookupEnvOrString("DB_LOCATION", "ld_db"), "folder location where the database is situated")

//	logLevel   = flag.String("log-level", lookupEnvOrString("LOG_LEVEL", "INFO"), "configure logging level")
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func main() {
	flag.Parse()
	//	loggingLevel, err := log.ParseLevel(*logLevel)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.SetLevel(loggingLevel)
	var handleRequests chan RoutingRequest = make(chan RoutingRequest)
	mux := NewInternalLdMux(handleRequests)

	//orchestrate the admin page and Server...

	_ = &http.Server{Addr: "localhost:8080", Handler: mux}
}
