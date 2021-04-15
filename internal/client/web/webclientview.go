// +build webclient

package main

import (
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"
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

func (cm *ClientModel) viewSetMatchControls() {
	usernameInput := cm.document.Call(
		"getElementById", "username")
	addClass(usernameInput, "hidden")
	matchButton := cm.document.Call(
		"getElementById", "beginMatchmakingButton")
	addClass(matchButton, "hidden")
	resignButton := cm.document.Call(
		"getElementById", "resignButton")
	removeClass(resignButton, "hidden")
	drawButton := cm.document.Call(
		"getElementById", "drawButton")
	removeClass(drawButton, "hidden")
}

func (cm *ClientModel) viewSetMatchMakingControls() {
	usernameInput := cm.document.Call(
		"getElementById", "username")
	removeClass(usernameInput, "hidden")
	matchButton := cm.document.Call(
		"getElementById", "beginMatchmakingButton")
	removeClass(matchButton, "hidden")
	resignButton := cm.document.Call(
		"getElementById", "resignButton")
	addClass(resignButton, "hidden")
	drawButton := cm.document.Call(
		"getElementById", "drawButton")
	addClass(drawButton, "hidden")
}

func (cm *ClientModel) viewSetMatchDetails() {
	opponentMatchDetailsName := cm.document.Call(
		"getElementById", "matchdetails_opponent_name")
	opponentMatchDetailsName.Set("innerText", cm.remoteMatchModel.opponentName)
	playerMatchDetailsName := cm.document.Call(
		"getElementById", "matchdetails_player_name")
	playerMatchDetailsName.Set("innerText", cm.playerName)
	opponentMatchDetailsRemainingTime := cm.document.Call(
		"getElementById", "matchdetails_opponent_remainingtime")
	cm.viewSetMatchDetailsPoints(cm.GetPlayerColor(),
		"matchdetails_player_points")
	cm.viewSetMatchDetailsPoints(cm.GetOpponentColor(),
		"matchdetails_opponent_points")
	opponentRemainingMs := cm.GetMaxTimeMs() -
		cm.GetPlayerElapsedMs(cm.GetOpponentColor())
	if opponentRemainingMs < 0 {
		opponentRemainingMs = 0
	}
	opponentMatchDetailsRemainingTime.Set("innerText",
		cm.formatTime(opponentRemainingMs))
	playerMatchDetailsRemainingTime := cm.document.Call(
		"getElementById", "matchdetails_player_remainingtime")
	playerRemainingMs := cm.GetMaxTimeMs() -
		cm.GetPlayerElapsedMs(cm.GetPlayerColor())
	if playerRemainingMs < 0 {
		playerRemainingMs = 0
	}
	playerMatchDetailsRemainingTime.Set("innerText",
		cm.formatTime(playerRemainingMs))
	drawButtonText := "Draw"
	if cm.GetRequestedDraw(cm.GetOpponentColor()) {
		drawButtonText = fmt.Sprintf("Draw, %s requested a draw",
			cm.remoteMatchModel.opponentName)
	} else if cm.GetRequestedDraw(cm.GetPlayerColor()) {
		drawButtonText = fmt.Sprintf("Draw, requesting a draw...")
	}
	drawButton := cm.document.Call("getElementById", "drawButton")
	drawButton.Set("innerText", drawButtonText)
}

func (cm *ClientModel) viewClearMatchDetails() {
	cm.document.Call("getElementById",
		"matchdetails_player_name").Set("innerText", "")
	cm.document.Call("getElementById",
		"matchdetails_opponent_name").Set("innerText", "")
	cm.document.Call("getElementById",
		"matchdetails_player_remainingtime").Set("innerText", "")
	cm.document.Call("getElementById",
		"matchdetails_opponent_remainingtime").Set("innerText", "")
	cm.document.Call("getElementById",
		"matchdetails_player_points").Set("innerText", "")
	cm.document.Call("getElementById",
		"matchdetails_opponent_points").Set("innerText", "")
}

func (cm *ClientModel) viewSetMatchDetailsPoints(
	color model.Color, elementID string) {
	points := cm.GetPointAdvantage(color)
	pointSummary := ""
	if points > 0 {
		pointSummary = strconv.Itoa(int(points))
	}
	matchDetailsPoints := cm.document.Call("getElementById", elementID)
	matchDetailsPoints.Set("innerText", pointSummary)
}

func (cm *ClientModel) formatTime(ms int64) string {
	return (time.Duration(ms) * time.Millisecond).String()
}

func (cm *ClientModel) viewSetGameOver(winner, winType string) {
	addClass(cm.document.Call("getElementById", "gameover_modal"), "gameover_modal")
	removeClass(cm.document.Call("getElementById", "gameover_modal"), "hidden")
	cm.document.Call("getElementById", "gameover_modal_text").Set("innerText",
		fmt.Sprintf("Winner: %s by %s", winner, winType))
}

func (clientModel *ClientModel) viewClearBoard() {
	elements := clientModel.document.Call("getElementsByClassName", "piece")
	for i := elements.Length() - 1; i >= 0; i-- {
		elements.Index(i).Call("remove")
	}
}

func (clientModel *ClientModel) viewInitBoard(playerColor model.Color) {
	for _, file := range clientModel.game.GetBoard() {
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
				div.Call("addEventListener", "touchstart",
					clientModel.genTouchStart(), false)
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
	if moveRequest.PromoteTo != nil {
		elMoving.Get("classList").Call("remove", "bp")
		elMoving.Get("classList").Call("remove", "wp")
		elMoving.Get("classList").Call("add",
			cm.GetPiece(newPos).ClientString())
	}
}

func (cm *ClientModel) viewHandleEnPassant(
	move model.Move, newPos model.Position, targetEmpty bool) {
	pawn := cm.game.GetBoard().Piece(newPos)
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
	king := cm.game.GetBoard().Piece(newPos)
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
	element.Get("classList").Call("remove", class)
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
