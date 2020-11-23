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

func (clientModel *ClientModel) viewClearBoard() {
	elements := clientModel.document.Call("getElementsByClassName", "piece")
	for i := elements.Length() - 1; i >= 0; i-- {
		elements.Index(i).Call("remove")
	}
}

func (clientModel *ClientModel) viewInitBoard(playerColor model.Color) {
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

func (cm *ClientModel) viewHandleMove(
	moveRequest model.MoveRequest, newPos model.Position, elMoving js.Value) {
	originalPositionClass :=
		getPositionClass(moveRequest.Position, cm.playerColor)
	newPositionClass := getPositionClass(newPos, cm.playerColor)
	elements := cm.document.Call("getElementsByClassName", newPositionClass)
	elementsLength := elements.Length()
	for i := elementsLength - 1; i >= 0; i-- {
		elements.Index(i).Call("remove")
	}
	cm.viewHandleCastle(moveRequest.Move, newPos)
	cm.viewHandleEnPassant(moveRequest.Move, newPos, elementsLength == 0)
	elMoving.Get("classList").Call("remove", originalPositionClass)
	elMoving.Get("classList").Call("add", newPositionClass)
}

func (cm *ClientModel) viewHandleEnPassant(
	move model.Move, newPos model.Position, targetEmpty bool) {
	pawn := cm.game.Board().Piece(newPos)
	if pawn.PieceType() == model.Pawn && move.X != 0 && targetEmpty {
		capturedY := pawn.Rank() + 1
		if move.Y > 0 {
			capturedY = pawn.Rank() - 1
		}
		capturedPosition := model.Position{pawn.File(), capturedY}
		capturedPosClass := getPositionClass(capturedPosition, cm.playerColor)
		elements := cm.document.Call("getElementsByClassName", capturedPosClass)
		elementsLength := elements.Length()
		for i := 0; i < elementsLength; i++ {
			elements.Index(i).Call("remove")
		}
	}
}

func (cm *ClientModel) viewHandleCastle(
	move model.Move, newPos model.Position) {
	king := cm.game.Board().Piece(newPos)
	if king.PieceType() == model.King &&
		(move.X < -1 || move.X > 1) {
		var rookPosition model.Position
		var rookPosClass string
		var rookNewPosClass string
		if king.File() == 2 {
			rookPosition = model.Position{0, king.Rank()}
			rookPosClass = getPositionClass(rookPosition, cm.playerColor)
			newRookPosition := model.Position{3, king.Rank()}
			rookNewPosClass = getPositionClass(newRookPosition, cm.playerColor)
		} else if king.File() == 6 {
			rookPosition = model.Position{7, king.Rank()}
			rookPosClass = getPositionClass(rookPosition, cm.playerColor)
			newRookPosition := model.Position{5, king.Rank()}
			rookNewPosClass = getPositionClass(newRookPosition, cm.playerColor)
		} else {
			panic("King not in valid castled position")
		}
		elements := cm.document.Call("getElementsByClassName", rookPosClass)
		elementsLength := elements.Length()
		for i := 0; i < elementsLength; i++ {
			elements.Index(i).Get("classList").Call("add", rookNewPosClass)
			elements.Index(i).Get("classList").Call("remove", rookPosClass)
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
