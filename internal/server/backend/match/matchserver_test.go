package matchserver

import (
	"strconv"
	"testing"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"
)

func TestMatchingServer(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
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
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
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
	response := <-player1.ResponseChanAsync
	response = <-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	response = <-player1.ResponseChanAsync
	if !liveMatch.Game.GameOver() || !response.GameOver || !response.Timeout {
		t.Error("Expected timed out game got", response)
	}
}

func TestMatchingServerInsufficientMaterialDraw(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	generator := func(black *Player, white *Player) Match {
		return newMatchNoPawns(black, white, 50)
	}
	matchingServer.StartCustomMatchServers(1, generator, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	response := <-player1.ResponseChanAsync
	response = <-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 0},
		Move:      model.Move{X: 0, Y: 7},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 7, Rank: 7},
		Move:      model.Move{X: 0, Y: -7},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 7, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 6, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 0},
		Move:      model.Move{X: 0, Y: 1},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 5, Rank: 0},
		Move:      model.Move{X: -2, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 2, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 7},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 7},
		Move:      model.Move{X: 2, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 5, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 2, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 6, Rank: 7},
		Move:      model.Move{X: -5, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 0},
		Move:      model.Move{X: 0, Y: 7},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 1},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 7},
		Move:      model.Move{X: 0, Y: -7},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 1},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 0},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 2, Rank: 1},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	black.GetSyncUpdate()
	response = <-white.ResponseChanAsync
	if !liveMatch.Game.GameOver() || !response.GameOver ||
		!response.Draw || response.Timeout {
		t.Error("Expected draw got", response)
	}
}

func TestMatchingServerInsufficientMaterialTimeoutDraw(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	generator := func(black *Player, white *Player) Match {
		return newMatchNoPawns(black, white, 50)
	}
	matchingServer.StartCustomMatchServers(1, generator, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	response := <-player1.ResponseChanAsync
	response = <-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 0},
		Move:      model.Move{X: 0, Y: 7},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 7, Rank: 7},
		Move:      model.Move{X: 0, Y: -7},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 7, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 6, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 0},
		Move:      model.Move{X: 0, Y: 1},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 5, Rank: 0},
		Move:      model.Move{X: -2, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 2, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 7},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 7},
		Move:      model.Move{X: 2, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 5, Rank: 7},
		Move:      model.Move{X: 1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 2, Rank: 0},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 6, Rank: 7},
		Move:      model.Move{X: -5, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 0},
		Move:      model.Move{X: 0, Y: 7},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 1},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 1, Rank: 7},
		Move:      model.Move{X: 0, Y: -7},
		PromoteTo: nil})
	white.GetSyncUpdate()
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 1},
		Move:      model.Move{X: -1, Y: 0},
		PromoteTo: nil})
	black.GetSyncUpdate()
	response = <-white.ResponseChanAsync
	if !liveMatch.Game.GameOver() || !response.GameOver ||
		!response.Draw || !response.Timeout {
		t.Error("Expected draw got", response)
	}
}

func TestMatchingServerDraw(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	<-player1.ResponseChanAsync
	<-player2.ResponseChanAsync
	player1.RequestChanAsync <- RequestAsync{RequestToDraw: true}
	response := <-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	if liveMatch.GetRequestedDraw() != player1 {
		t.Error("expected player1 to have requested a draw",
			liveMatch.GetRequestedDraw())
	}
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 1},
		Move:      model.Move{X: 0, Y: 2},
		PromoteTo: nil})
	if liveMatch.GetRequestedDraw() != nil {
		t.Error("expected move to have reset the request to draw",
			liveMatch.GetRequestedDraw())
	}
	player1.RequestChanAsync <- RequestAsync{RequestToDraw: true}
	<-player2.ResponseChanAsync
	player1.RequestChanAsync <- RequestAsync{RequestToDraw: true}
	<-player2.ResponseChanAsync
	tries := 0
	for liveMatch.GetRequestedDraw() != nil && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	if liveMatch.GetRequestedDraw() != nil {
		t.Error("Expected player1 to have toggled RequestToDraw",
			liveMatch.GetRequestedDraw())
	}
	player1.RequestChanAsync <- RequestAsync{RequestToDraw: true}
	response = <-player2.ResponseChanAsync
	if response.RequestToDraw != true || liveMatch.GetRequestedDraw() != player1 {
		t.Error("Expected player2 to receive a RequestToDraw",
			response)
	}
	player2.RequestChanAsync <- RequestAsync{RequestToDraw: true}
	response = <-player2.ResponseChanAsync
	if !liveMatch.Game.GameOver() || !response.GameOver ||
		!response.Draw {
		t.Error("Expected a draw got ", response.Draw, response.GameOver, liveMatch.Game.GameOver())
	}
}

func TestMatchingServerResignation(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	<-player1.ResponseChanAsync
	<-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	player1.RequestChanAsync <- RequestAsync{Resign: true}
	response := <-player1.ResponseChanAsync
	if !liveMatch.Game.GameOver() || !response.GameOver ||
		!response.Resignation {
		t.Error("Expected resignation got ", response)
	}
}

func TestMatchingServerPlayerSecondGame(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	go matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	<-player1.ResponseChanAsync
	<-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	player1.RequestChanAsync <- RequestAsync{Resign: true}
	response := <-player1.ResponseChanAsync
	player1.ClientDoneWithMatch()
	player2.ClientDoneWithMatch()
	tries = 0
	for len(matchingServer.LiveMatches()) == 1 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	if !liveMatch.Game.GameOver() || !response.GameOver ||
		!response.Resignation || len(matchingServer.LiveMatches()) > 0 {
		t.Error("Expected resignation got ", response.GameOver)
	}
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	<-player1.ResponseChanAsync
	<-player2.ResponseChanAsync
	liveMatch = matchingServer.LiveMatches()[0]
	player1.RequestChanAsync <- RequestAsync{Resign: true}
	response = <-player1.ResponseChanAsync
	if !liveMatch.Game.GameOver() || !response.GameOver ||
		!response.Resignation {
		t.Error("Expected resignation got ", response)
	}
	exitChan <- true
}

func TestMatchingServerValidMoves(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	<-player1.ResponseChanAsync
	<-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 1},
		Move:      model.Move{X: 0, Y: 2},
		PromoteTo: nil})
	response :=
		black.MakeMove(model.MoveRequest{
			Position:  model.Position{File: 3, Rank: 6},
			Move:      model.Move{X: 0, Y: -2},
			PromoteTo: nil})
	if !response {
		t.Error("Expected a valid move got", response)
	}
}

func TestMatchingServerInvalidMoves(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	<-player1.ResponseChanAsync
	<-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 1},
		Move:      model.Move{X: 0, Y: 2},
		PromoteTo: nil})
	response := black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 6},
		Move:      model.Move{X: 0, Y: -3},
		PromoteTo: nil})
	if response {
		t.Error("Expected a invalid move got", response)
	}
	response = black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 6},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	if !response {
		t.Error("Expected a valid move got", response)
	}
}

func TestMatchingServerCheckmate(t *testing.T) {
	player1 := NewPlayer("player1")
	player2 := NewPlayer("player2")
	matchingServer := NewMatchingServer()
	go matchingServer.MatchPlayer(player1)
	go matchingServer.MatchPlayer(player2)
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.StartMatchServers(1, exitChan)
	tries := 0
	for len(matchingServer.LiveMatches()) == 0 && tries < 10 {
		time.Sleep(time.Millisecond)
		tries++
	}
	response := <-player1.ResponseChanAsync
	response = <-player2.ResponseChanAsync
	liveMatch := matchingServer.LiveMatches()[0]
	black := liveMatch.black
	white := liveMatch.white
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 1},
		Move:      model.Move{X: 0, Y: 2},
		PromoteTo: nil})
	opponentMove := black.GetSyncUpdate()
	expectedMove := model.MoveRequest{
		Position:  model.Position{File: 4, Rank: 1},
		Move:      model.Move{X: 0, Y: 2},
		PromoteTo: nil}
	if *opponentMove != expectedMove {
		t.Error("Expected opponent's move got ", opponentMove)
	}
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 6},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	opponentMove = white.GetSyncUpdate()
	expectedMove = model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 6},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil}
	if *opponentMove != expectedMove {
		t.Error("Expected opponent's move got ", opponentMove)
	}
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 3, Rank: 0},
		Move:      model.Move{X: 4, Y: 4},
		PromoteTo: nil})
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 5},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 5, Rank: 0},
		Move:      model.Move{X: -3, Y: 3},
		PromoteTo: nil})
	black.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 0, Rank: 4},
		Move:      model.Move{X: 0, Y: -1},
		PromoteTo: nil})
	white.MakeMove(model.MoveRequest{
		Position:  model.Position{File: 7, Rank: 4},
		Move:      model.Move{X: -2, Y: 2},
		PromoteTo: nil})
	if !liveMatch.Game.GameOver() {
		t.Error("Expected gameover got ", liveMatch)
	}
	response = <-black.ResponseChanAsync
	if !response.GameOver || !(response.Winner == white.name) {
		t.Error("Expected checkmate got ", response)
	}
}

func TestMatchingServerEngineTimeout(t *testing.T) {
	matchingServer := NewMatchingServerWithEngine("localhost:50000",
		time.Millisecond, time.Millisecond)
	if matchingServer.botMatchingEnabled {
		t.Error("Expected bot matching to be disabled due to failed connection",
			"with engine..")
	}
}

func TestMatchingServerMultiple(t *testing.T) {
	matchingServer := NewMatchingServer()
	players := []*Player{}
	for i := 0; i < 7; i++ {
		players = append(players, NewPlayer("player"+strconv.Itoa(i)))
		go matchingServer.MatchPlayer(players[i])
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
