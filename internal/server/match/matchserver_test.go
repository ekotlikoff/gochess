package matchserver

import (
	"github.com/Ekotlikoff/gochess/internal/model"
	"strconv"
	"testing"
	"time"
)

func TestMatchingServer(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	for len(matchingServer.LiveMatches()) == 0 {
	}
	liveMatch := matchingServer.LiveMatches()[0]
	if liveMatch.black.name != "player1" && liveMatch.white.name != "player1" {
		t.Error("Expected our players got ", liveMatch.white.name)
	}
}

func TestMatchingServerTimeout(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	generator := func(black *Player, white *Player) Match {
		return NewMatch(black, white, 50)
	}
	matchingServer.StartCustomMatchServers(1, generator, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	response := <-player1.responseChanAsync
	if !liveMatch.game.GameOver() || !response.GameOver || !response.Timeout {
		t.Error("Expected timed out game got", response)
	}
}

func TestMatchingServerDraw(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	player1.requestChanAsync <- RequestAsync{RequestToDraw: true}
	response := <-player2.responseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	if liveMatch.requestedDraw != &player1 {
		t.Error("Expected player1 to have requested a draw",
			liveMatch.requestedDraw)
	}
	player1.requestChanAsync <- RequestAsync{RequestToDraw: true}
	tries := 0
	for liveMatch.requestedDraw != nil && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	if liveMatch.requestedDraw != nil {
		t.Error("Expected player1 to have toggled RequestToDraw",
			liveMatch.requestedDraw)
	}
	player1.requestChanAsync <- RequestAsync{RequestToDraw: true}
	response = <-player2.responseChanAsync
	if response.RequestToDraw != true || liveMatch.requestedDraw != &player1 {
		t.Error("Expected player2 to receive a RequestToDraw",
			response)
	}
	player2.requestChanAsync <- RequestAsync{RequestToDraw: true}
	response = <-player2.responseChanAsync
	if !liveMatch.game.GameOver() || !response.GameOver ||
		!response.Draw {
		t.Error("Expected a draw got ", response.Draw, response.GameOver, liveMatch.game.GameOver())
	}
}

func TestMatchingServerResignation(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	player1.requestChanAsync <- RequestAsync{Resign: true}
	response := <-player1.responseChanAsync
	if !liveMatch.game.GameOver() || !response.GameOver ||
		!response.Resignation {
		t.Error("Expected resignation got ", response)
	}
}

func TestMatchingServerValidMoves(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{model.Position{3, 1}, model.Move{0, 2}})
	response :=
		black.MakeMove(model.MoveRequest{model.Position{3, 6}, model.Move{0, -2}})
	if !response {
		t.Error("Expected a valid move got", response)
	}
}

func TestMatchingServerInvalidMoves(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{model.Position{3, 1}, model.Move{0, 2}})
	response := black.MakeMove(model.MoveRequest{model.Position{3, 6}, model.Move{0, -3}})
	if response {
		t.Error("Expected a invalid move got", response)
	}
	response = black.MakeMove(model.MoveRequest{model.Position{3, 6}, model.Move{0, -1}})
	if !response {
		t.Error("Expected a valid move got", response)
	}
}

func TestMatchingServerCheckmate(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(&player1)
	go matchingServer.MatchPlayer(&player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{model.Position{4, 1}, model.Move{0, 2}})
	opponentMove := black.GetSyncUpdate()
	expectedMove := model.MoveRequest{model.Position{4, 1}, model.Move{0, 2}}
	if opponentMove != expectedMove {
		t.Error("Expected opponent's move got ", opponentMove)
	}
	black.MakeMove(model.MoveRequest{model.Position{0, 6}, model.Move{0, -1}})
	opponentMove = white.GetSyncUpdate()
	expectedMove = model.MoveRequest{model.Position{0, 6}, model.Move{0, -1}}
	if opponentMove != expectedMove {
		t.Error("Expected opponent's move got ", opponentMove)
	}
	white.MakeMove(model.MoveRequest{model.Position{3, 0}, model.Move{4, 4}})
	black.MakeMove(model.MoveRequest{model.Position{0, 5}, model.Move{0, -1}})
	white.MakeMove(model.MoveRequest{model.Position{5, 0}, model.Move{-3, 3}})
	black.MakeMove(model.MoveRequest{model.Position{0, 4}, model.Move{0, -1}})
	white.MakeMove(model.MoveRequest{model.Position{7, 4}, model.Move{-2, 2}})
	if !liveMatch.game.GameOver() {
		t.Error("Expected gameover got ", liveMatch)
	}
	response := <-black.responseChanAsync
	if !response.GameOver || !(response.Winner == white.name) {
		t.Error("Expected checkmate got ", response)
	}
}

func TestMatchingServerMultiple(t *testing.T) {
	matchingServer := NewMatchingServer()
	players := []Player{}
	for i := 0; i < 7; i++ {
		players = append(players, NewPlayer("player"+strconv.Itoa(i)))
		go matchingServer.MatchPlayer(&players[i])
	}
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(5, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) != 3 && tries < 50 {
		time.Sleep(time.Millisecond)
		tries++
	}
	if len(matchingServer.LiveMatches()) != 3 {
		t.Error("Expected all players matched got ",
			len(matchingServer.LiveMatches()))
	}
}
