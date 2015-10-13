package main

import (
	"flag"
	"log"

	"github.com/joyrexus/xhub"
)

var (
	addr   string
	dbfile string
)

func main() {
	flag.StringVar(&addr, "addr", "localhost:8081", "host name or ip address")
	flag.StringVar(&dbfile, "dbfile", "xhub.db", "path to database file")
	flag.Parse()

	srv := xhub.NewServer(addr, dbfile)
	log.Fatal(srv.ListenAndServe())
}
