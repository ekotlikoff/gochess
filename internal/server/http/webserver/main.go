package main

import (
	"gochess/internal/model"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"syscall/js"
)

type GameType uint8

const (
	Local  = GameType(iota)
	Remote = GameType(iota)
)

type ClientModel struct {
	gameType                 GameType
	game                     *model.Game
	playerColor              model.Color
	elDragging               js.Value
	isMouseDown              bool
	pieceDragging            *model.Piece
	positionOriginal         model.Position
	document                 js.Value
	board                    js.Value
	draggingOrigTransform    js.Value
	isMatchmaking, isMatched bool
	matchingServerURI        string
	client                   *http.Client
	mutex                    sync.Mutex
}

func main() {
	done := make(chan struct{}, 0)
	game := model.NewGame()
	jar, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	clientModel := ClientModel{
		gameType: Local, game: &game, playerColor: model.White,
		document: js.Global().Get("document"),
		board: js.Global().Get("document").Call(
			"getElementById", "board-layout-chessboard"),
		matchingServerURI: "http://localhost:8081",
		client:            client,
	}
	clientModel.initListeners()
	// TODO make http calls to interact with server
	clientModel.initStyle()
	clientModel.initBoard(model.White)
	<-done
}
