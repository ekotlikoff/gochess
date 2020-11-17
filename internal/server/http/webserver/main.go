package main

import (
	"gochess/internal/model"
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
	draggingOrigTransform    js.Value
	document                 js.Value
	board                    js.Value
	isMatchmaking, isMatched bool
	matchingServerURI        string
}

func main() {
	done := make(chan struct{}, 0)
	game := model.NewGame()
	clientModel := ClientModel{
		gameType: Local, game: &game, playerColor: model.White,
		document: js.Global().Get("document"),
		board: js.Global().Get("document").Call(
			"getElementById", "board-layout-chessboard"),
		matchingServerURI: "http://192.168.1.243:8081",
	}
	clientModel.initListeners()
	// TODO make http calls to interact with server
	clientModel.initStyle()
	clientModel.initBoard(model.White)
	<-done
}
