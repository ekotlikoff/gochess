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
	gameType              GameType
	game                  *model.Game
	playerColor           model.Color
	elDragging            js.Value
	isMouseDown           bool
	pieceDragging         *model.Piece
	positionOriginal      model.Position
	draggingOrigTransform js.Value
	document              js.Value
	board                 js.Value
}

func main() {
	done := make(chan struct{}, 0)
	clientModel := ClientModel{}
	clientModel.gameType = Local
	game := model.NewGame()
	clientModel.game = &game
	clientModel.playerColor = model.White
	clientModel.document = js.Global().Get("document")
	clientModel.board =
		clientModel.document.Call("getElementById", "board-layout-chessboard")
	// TODO make http calls to interact with server
	initStyle(clientModel.document)
	clientModel.document.Call("addEventListener", "mousemove",
		genMouseMove(&clientModel), false)
	clientModel.document.Call("addEventListener", "mouseup",
		genMouseUp(&clientModel), false)
	clientModel.board.Call("addEventListener", "contextmenu", js.FuncOf(preventDefault), false)
	initBoard(&clientModel, model.White)
	<-done
}
