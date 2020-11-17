package main

import (
	"fmt"
	"gochess/internal/model"
	"strconv"
	"syscall/js"
)

func (clientModel *ClientModel) initStyle() {
	document := clientModel.document
	styleEl := document.Call("createElement", "style")
	document.Get("head").Call("appendChild", styleEl)
	styleSheet := styleEl.Get("sheet")
	for x := 1; x < 9; x++ {
		for y := 1; y < 9; y++ {
			selector := ".square-" + strconv.Itoa(x*10+y)
			transform := fmt.Sprintf(
				"{transform: translate(%d%%,%d%%);}",
				x*100-100, 700-((y-1)*100),
			)
			styleSheet.Call("insertRule", selector+transform)
		}
	}
}

func (clientModel *ClientModel) resetBoard() {
	elements := clientModel.document.Call("getElementsByClassName", "piece")
	for i := 0; i < elements.Length(); i++ {
		elements.Index(i).Call("remove")
	}
}

func (clientModel *ClientModel) initBoard(playerColor model.Color) {
	for _, file := range clientModel.game.Board() {
		for _, piece := range file {
			if piece != nil {
				div := clientModel.document.Call("createElement", "div")
				div.Get("classList").Call("add", "piece")
				div.Get("classList").Call("add", piece.ClientString())
				div.Get("classList").Call(
					"add", getPositionClass(piece.Position(), playerColor))
				clientModel.board.Call("appendChild", div)
				div.Call("addEventListener", "mousedown",
					clientModel.genMouseDown(), false)
			}
		}
	}
}

func (clientModel *ClientModel) buttonBeginLoading(button js.Value) js.Value {
	i := clientModel.document.Call("createElement", "div")
	i.Get("classList").Call("add", "loading")
	button.Call("appendChild", i)
	return i
}

func addClass(element js.Value, class string) {
	element.Get("classList").Call("add", class)
}

func removeClass(element js.Value, class string) {
	element.Get("classList").Call("remove", "dragging")
}

func (clientModel *ClientModel) viewDragPiece(
	elDragging js.Value, event js.Value) {
	x, y, squareWidth, squareHeight, _, _ :=
		clientModel.getEventMousePosition(event)
	pieceWidth := elDragging.Get("clientWidth").Float()
	pieceHeight := elDragging.Get("clientHeight").Float()
	percentX := 100 * (float64(x) - pieceWidth/2) / float64(squareWidth)
	percentY := 100 * (float64(y) - pieceHeight/2) / float64(squareHeight)
	elDragging.Get("style").Set("transform",
		"translate("+fmt.Sprintf("%f%%,%f%%)", percentX, percentY))
}

func getPositionClass(position model.Position, playerColor model.Color) string {
	class := "square-"
	if playerColor == model.Black {
		class += strconv.Itoa(int(8 - position.File))
		class += strconv.Itoa(int(8 - position.Rank))
	} else {
		class += strconv.Itoa(int(position.File + 1))
		class += strconv.Itoa(int(position.Rank + 1))
	}
	return class
}
