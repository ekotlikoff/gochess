// +build ignore

package main

import (
	"gochess/internal/server/http/webserver"
	"gochess/internal/server/match"
)

func main() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	go matchingServer.StartMatchServers(10, exitChan)
	println("Listening on port 8000...")
	webserver.Serve(&matchingServer, 8000, nil, false)
}
