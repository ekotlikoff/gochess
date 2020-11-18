package main

import (
	"fmt"
	"gochess/internal/model"
	"net/http"
	"syscall/js"
)

func (clientModel *ClientModel) initListeners() {
	clientModel.document.Call("addEventListener", "mousemove",
		clientModel.genMouseMove(), false)
	clientModel.document.Call("addEventListener", "mouseup",
		clientModel.genMouseUp(), false)
	js.Global().Set("beginMatchmaking", clientModel.genBeginMatchmaking())
	clientModel.board.Call("addEventListener", "contextmenu", js.FuncOf(preventDefault), false)
}

func preventDefault(this js.Value, i []js.Value) interface{} {
	if len(i) > 0 {
		i[0].Call("preventDefault")
	}
	return 0
}

func (clientModel *ClientModel) genMouseDown() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if len(i) > 0 && !clientModel.isMouseDown {
			clientModel.isMouseDown = true
			i[0].Call("preventDefault")
			clientModel.elDragging = this
			_, _, _, _, gridX, gridY :=
				clientModel.getEventMousePosition(i[0])
			clientModel.positionOriginal =
				model.Position{uint8(gridX), uint8(7 - gridY)}
			clientModel.pieceDragging = clientModel.game.Board()[gridX][7-gridY]
			addClass(clientModel.elDragging, "dragging")
			clientModel.draggingOrigTransform =
				clientModel.elDragging.Get("style").Get("transform")
		}
		return 0
	})
}

func (clientModel *ClientModel) genMouseMove() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		i[0].Call("preventDefault")
		if clientModel.isMouseDown {
			clientModel.viewDragPiece(clientModel.elDragging, i[0])
		}
		return 0
	})
}

func (clientModel *ClientModel) genMouseUp() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		cm := clientModel
		if cm.isMouseDown && len(i) > 0 {
			i[0].Call("preventDefault")
			cm.elDragging.Get("style").Set("transform", cm.draggingOrigTransform)
			removeClass(cm.elDragging, "dragging")
			originalPositionClass :=
				getPositionClass(cm.positionOriginal, cm.playerColor)
			_, _, _, _, gridX, gridY := clientModel.getEventMousePosition(i[0])
			positionDragging := model.Position{uint8(gridX), uint8(7 - gridY)}
			move := model.Move{
				int8(positionDragging.File) - int8(cm.positionOriginal.File),
				int8(positionDragging.Rank) - int8(cm.positionOriginal.Rank),
			}
			err := cm.game.Move(model.MoveRequest{cm.positionOriginal, move})
			fmt.Println(err)
			if err == nil && (cm.gameType == Local ||
				(cm.gameType == Remote && cm.playerColor == cm.game.Turn())) {
				newPositionClass :=
					getPositionClass(positionDragging, cm.playerColor)
				elements := cm.document.Call("getElementsByClassName", newPositionClass)
				elementsLength := elements.Length()
				for i := 0; i < elementsLength; i++ {
					elements.Index(i).Call("remove")
				}
				cm.elDragging.Get("classList").Call("remove", originalPositionClass)
				cm.elDragging.Get("classList").Call("add", newPositionClass)
				cm.handleCastle(move)
				cm.handleEnPassant(move, elementsLength == 0)
			}
			cm.elDragging = js.Undefined()
			cm.isMouseDown = false
		}
		return 0
	})
}

func (clientModel *ClientModel) genBeginMatchmaking() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if !clientModel.isMatchmaking && !clientModel.isMatched {
			go lookForMatch()
		}
		return 0
	})
}

func (clientModel *ClientModel) lookForMatch() js.Func {
	clientModel.mutex.Lock()
	clientModel.isMatchmaking = true
	clientModel.mutex.Unlock()
	buttonLoader := clientModel.buttonBeginLoading(
		clientModel.document.Call(
			"getElementById", "beginMatchmakingButton"))
	// TODO
	requestBody, err := json.Marshal(map[string]string{
		"username": "my_username",
	})
	resp, err := http.Post(clientModel.matchingServerURI, "application/json",
		bytes.NewBuffer(requestBody))
	// - post to clientModel.matchingServerURI with username to begin session
	// - Store state in clientModel to remember that the session is started
	// - GET /match to begin matching
	// - Once matched stops displaying loading icon and briefly displays matched icon
	buttonLoader.Call("remove")
	// - Once matched reset board and set player color and time remaining
}

func (clientModel *ClientModel) getEventMousePosition(event js.Value) (
	int, int, int, int, int, int) {
	rect := clientModel.board.Call("getBoundingClientRect")
	width := rect.Get("right").Int() - rect.Get("left").Int()
	height := rect.Get("bottom").Int() - rect.Get("top").Int()
	squareWidth := width / 8
	squareHeight := height / 8
	x := event.Get("clientX").Int() - rect.Get("left").Int()
	gridX := x / squareWidth
	if x > width || gridX > 7 {
		x = width
		gridX = 7
	} else if x < 0 || gridX < 0 {
		x = 0
		gridX = 0
	}
	y := event.Get("clientY").Int() - rect.Get("top").Int()
	gridY := y / squareHeight
	if y > height || gridY > 7 {
		y = height
		gridY = 7
	} else if y < 0 || gridY < 0 {
		y = 0
		gridY = 0
	}
	return x, y, squareWidth, squareHeight, gridX, gridY
}

func (cm *ClientModel) handleEnPassant(move model.Move, targetEmpty bool) {
	pawn := cm.pieceDragging
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

func (cm *ClientModel) handleCastle(move model.Move) {
	king := cm.pieceDragging
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
