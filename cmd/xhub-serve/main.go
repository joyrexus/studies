package main

import (
	"fmt"
	"log"

	"github.com/joyrexus/xhub"
)

const (
	host   = "localhost"  // host name or ip address
	port   = 8081         // host port number
	dbfile = "xhub.db" // path to file to use for persisting xhub data
)

func main() {
	addr := fmt.Sprintf("%s:%d", host, port)
	srv := xhub.NewServer(addr, dbfile)
	log.Fatal(srv.ListenAndServe())
}
