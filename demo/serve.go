package main

import (
	"fmt"
	"log"
	// "net/http"

	"github.com/joyrexus/studies"
)

const (
	host   = "localhost"  // host name or ip address
	port   = 8081         // host port number
	dbfile = "studies.db" // path to file to use for persisting study data
)

func main() {
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := studies.NewServer(addr, dbfile)
	log.Fatal(srv.ListenAndServe())
}
