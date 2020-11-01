package server

import (
	"strconv"
	"testing"
	"time"
)

func TestMatchingServer(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 1, exitChan)
	for len(matchingServer.LiveMatches()) == 0 {
	}
	liveMatch := matchingServer.LiveMatches()[0]
	if liveMatch.black.name != "player1" && liveMatch.white.name != "player1" {
		t.Error("Expected our players got ", liveMatch.black.name)
	}
}

func TestMatchingServerResignation(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 1, exitChan)
	go func() { player1.requestChanAsync <- RequestAsync{resign: true} }()
	response := ResponseAsync{}
	select {
	case response = <-player1.responseChanAsync:
	}
	liveMatch := matchingServer.LiveMatches()[0]
	if !liveMatch.game.GameOver() || !response.gameOver ||
		!response.resignation {
		t.Error("Expected our players got ", liveMatch.black.name)
	}
}

func TestMatchingServerMultiple(t *testing.T) {
	players := []Player{}
	for i := 0; i < 7; i++ {
		players = append(players, NewPlayer("player"+strconv.Itoa(i)))
		matchingChan <- &players[i]
	}
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 5, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) != 3 && tries < 5 {
		time.Sleep(time.Millisecond * 10)
		tries++
	}
	if len(matchingServer.LiveMatches()) != 3 {
		t.Error("Expected all players matched got ", len(matchingServer.LiveMatches()))
	}
}
