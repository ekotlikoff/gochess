package server

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	fmt.Println(matchingServer.liveMatches)
	matchingServer.Serve(matchingChan, 1, exitChan)
	fmt.Println(matchingServer.liveMatches)
	go func() { player1.requestChan <- Request{resign: true} }()
	go func() { player2.requestChan <- Request{resign: true} }()
	fmt.Println(matchingServer.liveMatches)
	select {
	case <-player1.responseChan:
	case <-player2.responseChan:
	}
	fmt.Println(matchingServer.liveMatches)
	t.Error("Expected blah got ")
}
