package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochess/internal/model"
	"gochess/internal/server/http/webserver"
	"syscall/js"
	"time"
)

var ctp string = "application/json"

func (clientModel *ClientModel) initListeners() {
	clientModel.document.Call("addEventListener", "mousemove",
		clientModel.genMouseMove(), false)
	clientModel.document.Call("addEventListener", "mouseup",
		clientModel.genMouseUp(), false)
	clientModel.board.Call("addEventListener", "contextmenu",
		js.FuncOf(preventDefault), false)
	js.Global().Set("beginMatchmaking", clientModel.genBeginMatchmaking())
}

func (clientModel *ClientModel) genMouseDown() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if len(i) > 0 && !clientModel.isMouseDown {
			clientModel.isMouseDown = true
			i[0].Call("preventDefault")
			clientModel.elDragging = this
			_, _, _, _, gridX, gridY :=
				clientModel.getEventMousePosition(i[0])
			clientModel.positionOriginal = clientModel.getPositionFromGrid(
				uint8(gridX), uint8(gridY))
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
			cm.isMouseDown = false
			elDragging := cm.elDragging
			cm.elDragging = js.Undefined()
			i[0].Call("preventDefault")
			_, _, _, _, gridX, gridY := clientModel.getEventMousePosition(i[0])
			newPosition :=
				clientModel.getPositionFromGrid(uint8(gridX), uint8(gridY))
			moveRequest := model.MoveRequest{
				clientModel.positionOriginal,
				model.Move{
					int8(newPosition.File) - int8(cm.positionOriginal.File),
					int8(newPosition.Rank) - int8(cm.positionOriginal.Rank),
				},
			}
			if cm.gameType == Local || cm.playerColor == cm.game.Turn() {
				go func() {
					cm.takeMove(moveRequest, newPosition, elDragging)
					elDragging.Get("style").Set("transform", cm.draggingOrigTransform)
					removeClass(elDragging, "dragging")
				}()
			} else {
				elDragging.Get("style").Set("transform", cm.draggingOrigTransform)
				removeClass(elDragging, "dragging")
			}
		}
		return 0
	})
}

func (cm *ClientModel) takeMove(
	moveRequest model.MoveRequest, newPos model.Position, elMoving js.Value) {
	successfulRemoteMove := false
	if cm.gameType == Remote {
		movePayloadBuf := new(bytes.Buffer)
		json.NewEncoder(movePayloadBuf).Encode(moveRequest)
		resp, err := cm.client.Post(
			cm.matchingServerURI+"sync", ctp, movePayloadBuf)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return
			}
			successfulRemoteMove = true
		}
	}
	err := cm.game.Move(moveRequest)
	fmt.Println(err)
	if err == nil {
		cm.viewHandleMove(moveRequest, newPos, elMoving)
	} else {
		if successfulRemoteMove {
			// TODO handle strange case where http call was successful but local
			// game did not accept the move.
			println("FATAL: We do not expect an unsuccessful local move when remote succeeds")
		}
	}
}

func (cm *ClientModel) listenForOpponentMove(endRemoteGame chan bool) {
	// TODO need to write to the endRemoteGame chan when game is over
	for true {
		select {
		case <-endRemoteGame:
			return
		default:
		}
		resp, err := cm.client.Get(cm.matchingServerURI + "sync")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				continue
			}
			opponentMove := model.MoveRequest{}
			json.NewDecoder(resp.Body).Decode(&opponentMove)
			cm.mutex.Lock()
			err := cm.game.Move(opponentMove)
			cm.mutex.Unlock()
			if err != nil {
				println("FATAL: We do not expect an invalid move from the opponent")
			}
			newPos := model.Position{
				opponentMove.Position.File + uint8(opponentMove.Move.X),
				opponentMove.Position.Rank + uint8(opponentMove.Move.Y),
			}
			originalPosClass :=
				getPositionClass(opponentMove.Position, cm.playerColor)
			elements := cm.document.Call("getElementsByClassName", originalPosClass)
			elMoving := elements.Index(0)
			cm.viewHandleMove(opponentMove, newPos, elMoving)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (clientModel *ClientModel) genBeginMatchmaking() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if !clientModel.isMatchmaking && !clientModel.isMatched {
			go clientModel.lookForMatch()
		}
		return 0
	})
}

func (clientModel *ClientModel) lookForMatch() {
	clientModel.mutex.Lock()
	clientModel.isMatchmaking = true
	buttonLoader := clientModel.buttonBeginLoading(
		clientModel.document.Call(
			"getElementById", "beginMatchmakingButton"))
	clientModel.mutex.Unlock()
	if !clientModel.hasSession {
		username := clientModel.document.Call(
			"getElementById", "username").Get("value").String()
		credentialsBuf := new(bytes.Buffer)
		credentials := webserver.Credentials{username}
		json.NewEncoder(credentialsBuf).Encode(credentials)
		resp, err := clientModel.client.Post(
			clientModel.matchingServerURI+"session", ctp, credentialsBuf,
		)
		if err == nil {
			resp.Body.Close()
		}
		clientModel.playerName = username
		clientModel.hasSession = true
	}
	resp, err := clientModel.client.Get(clientModel.matchingServerURI + "match")
	if err == nil {
		var matchResponse webserver.MatchedResponse
		json.NewDecoder(resp.Body).Decode(&matchResponse)
		resp.Body.Close()
		clientModel.playerColor = matchResponse.Color
		clientModel.opponentName = matchResponse.OpponentName
		clientModel.resetGame()
		// - TODO once matched briefly display matched icon?
		// - TODO once matched set and display time remaining
		clientModel.gameType = Remote
		clientModel.isMatched = true
		clientModel.isMatchmaking = false
		buttonLoader.Call("remove")
		clientModel.endRemoteGameChan = make(chan bool, 0)
		clientModel.viewSetMatchDetails()
		go clientModel.listenForOpponentMove(clientModel.endRemoteGameChan)
	}
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

// To flip black to be on the bottom we do two things, everything is flipped
// in the view (see getPositionClass), and everything is flipped onClick here.
func (cm *ClientModel) getPositionFromGrid(
	gridX uint8, gridY uint8) model.Position {
	if cm.playerColor == model.White {
		return model.Position{uint8(gridX), uint8(7 - gridY)}
	} else {
		return model.Position{uint8(7 - gridX), uint8(gridY)}
	}
}

func preventDefault(this js.Value, i []js.Value) interface{} {
	if len(i) > 0 {
		i[0].Call("preventDefault")
	}
	return 0
}

func (clientModel *ClientModel) resetGame() {
	game := model.NewGame()
	clientModel.game = &game
	clientModel.viewClearBoard()
	clientModel.viewInitBoard(clientModel.playerColor)
}
