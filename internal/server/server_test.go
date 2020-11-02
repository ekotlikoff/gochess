package server

import (
	"gochess/internal/model"
	"strconv"
	"testing"
	"time"
)

func TestMatchingServer(t *testing.T) {
	matchingChan := make(chan *Player, 20)
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
		t.Error("Expected our players got ", liveMatch.white.name)
	}
}

func TestMatchingServerTimeout(t *testing.T) {
	matchingChan := make(chan *Player, 20)
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	generator := func(black *Player, white *Player) Match {
		return newMatch(black, white, 50)
	}
	matchingServer.ServeCustomMatch(matchingChan, 1, generator, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond * 5)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	response := ResponseAsync{}
	select {
	case response = <-player1.responseChanAsync:
	}
	if !liveMatch.game.GameOver() || !response.gameOver || !response.timeout {
		t.Error("Expected timed out game got", response)
	}
}

func TestMatchingServerResignation(t *testing.T) {
	matchingChan := make(chan *Player, 20)
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 5 {
		time.Sleep(time.Millisecond * 10)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	go func() { player1.requestChanAsync <- RequestAsync{resign: true} }()
	response := ResponseAsync{}
	select {
	case response = <-player1.responseChanAsync:
	}
	if !liveMatch.game.GameOver() || !response.gameOver ||
		!response.resignation {
		t.Error("Expected our players got ", liveMatch.black.name)
	}
}

func TestMatchingServerValidMoves(t *testing.T) {
	matchingChan := make(chan *Player, 20)
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 5 {
		time.Sleep(time.Millisecond * 10)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	makeMove(white, model.Position{3, 1}, model.Move{0, 2})
	response := makeMove(black, model.Position{3, 6}, model.Move{0, -2})
	if !response.moveSuccess {
		t.Error("Expected a valid move got", response)
	}
}

func TestMatchingServerInvalidMoves(t *testing.T) {
	matchingChan := make(chan *Player, 20)
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 5 {
		time.Sleep(time.Millisecond * 10)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	makeMove(white, model.Position{3, 1}, model.Move{0, 2})
	response := makeMove(black, model.Position{3, 6}, model.Move{0, -3})
	if response.moveSuccess {
		t.Error("Expected a invalid move got", response)
	}
	response = makeMove(black, model.Position{3, 6}, model.Move{0, -1})
	if !response.moveSuccess {
		t.Error("Expected a valid move got", response)
	}
}

func TestMatchingServerCheckmate(t *testing.T) {
	matchingChan := make(chan *Player, 20)
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingChan <- &player1
	matchingChan <- &player2
	matchingServer := NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(matchingChan, 1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 5 {
		time.Sleep(time.Millisecond * 10)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	makeMove(white, model.Position{4, 1}, model.Move{0, 2})
	makeMove(black, model.Position{0, 6}, model.Move{0, -1})
	makeMove(white, model.Position{3, 0}, model.Move{4, 4})
	makeMove(black, model.Position{0, 5}, model.Move{0, -1})
	makeMove(white, model.Position{5, 0}, model.Move{-3, 3})
	makeMove(black, model.Position{0, 4}, model.Move{0, -1})
	makeMove(white, model.Position{7, 4}, model.Move{-2, 2})
	if !liveMatch.game.GameOver() {
		t.Error("Expected gameover got ", liveMatch)
	}
	response := ResponseAsync{}
	select {
	case response = <-black.responseChanAsync:
	}
	if !response.gameOver || !(response.winner == white.name) {
		t.Error("Expected checkmate got ", response)
	}
}

func makeMove(
	player *Player, position model.Position, move model.Move,
) ResponseSync {
	player.requestChanSync <- RequestSync{position, move}
	return <-player.responseChanSync
}

func TestMatchingServerMultiple(t *testing.T) {
	matchingChan := make(chan *Player, 20)
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
