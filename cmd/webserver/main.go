package main

import (
	"github.com/Ekotlikoff/gochess/internal/server/api/http"
	"github.com/Ekotlikoff/gochess/internal/server/match"
)

func main() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	go matchingServer.StartMatchServers(10, exitChan)
	println("Listening on port 8000...")
	httpserver.Serve(&matchingServer, 8000, nil, false)
}
