// +build ignore
// TODO this should be moved to bin

package main

import (
	"flag"
	"gochess/internal/server/http/apiserver"
	"gochess/internal/server/match"
	"log"
	"net/http"
)

var (
	listen = flag.String("listen", ":8080", "listen address")
	dir    = flag.String("dir", ".", "directory to serve")
)

func main() {
	flag.Parse()
	log.Printf("starting apiserver on 8081...")
	startAPIServer()
	log.Printf("listening on %q...", *listen)
	log.Fatal(http.ListenAndServe(*listen, http.FileServer(http.Dir(*dir))))
}

func startAPIServer() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	go matchingServer.StartMatchServers(10, exitChan)
	go apiserver.Serve(&matchingServer, 8081, nil, true)
}
