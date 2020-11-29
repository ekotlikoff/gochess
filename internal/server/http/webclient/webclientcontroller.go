// +build webclient

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Ekotlikoff/gochess/internal/model"
	"github.com/Ekotlikoff/gochess/internal/server/http/webserver"
	"io/ioutil"
	"log"
	"syscall/js"
	"time"
)

var ctp string = "application/json"

func (clientModel *ClientModel) initController(quiet bool) {
	clientModel.document.Call("addEventListener", "mousemove",
		clientModel.genMouseMove(), false)
	clientModel.document.Call("addEventListener", "touchmove",
		clientModel.genTouchMove(), false)
	clientModel.document.Call("addEventListener", "mouseup",
		clientModel.genMouseUp(), false)
	clientModel.document.Call("addEventListener", "touchend",
		clientModel.genTouchEnd(), false)
	clientModel.board.Call("addEventListener", "contextmenu",
		js.FuncOf(preventDefault), false)
	js.Global().Set("beginMatchmaking", clientModel.genBeginMatchmaking())
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
}

func (clientModel *ClientModel) genMouseDown() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if len(i) > 0 && !clientModel.GetIsMouseDown() {
			i[0].Call("preventDefault")
			clientModel.handleClickStart(this, i[0])
		}
		return 0
	})
}

func (clientModel *ClientModel) genTouchStart() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if len(i) > 0 && !clientModel.GetIsMouseDown() {
			i[0].Call("preventDefault")
			touch := i[0].Get("touches").Index(0)
			clientModel.handleClickStart(this, touch)
		}
		return 0
	})
}

func (clientModel *ClientModel) handleClickStart(
	this js.Value, event js.Value) {
	clientModel.SetIsMouseDown(true)
	clientModel.SetDraggingElement(this)
	_, _, _, _, gridX, gridY :=
		clientModel.getEventMousePosition(event)
	clientModel.positionOriginal = clientModel.getPositionFromGrid(
		uint8(gridX), uint8(gridY))
	addClass(clientModel.GetDraggingElement(), "dragging")
	clientModel.SetDraggingOriginalTransform(
		clientModel.GetDraggingElement().Get("style").Get("transform"))
}

func (clientModel *ClientModel) genMouseMove() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		i[0].Call("preventDefault")
		clientModel.handleMoveEvent(i[0])
		return 0
	})
}

func (clientModel *ClientModel) genTouchMove() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		i[0].Call("preventDefault")
		touch := i[0].Get("touches").Index(0)
		clientModel.handleMoveEvent(touch)
		return 0
	})
}

func (clientModel *ClientModel) handleMoveEvent(moveEvent js.Value) {
	if clientModel.GetIsMouseDown() {
		clientModel.viewDragPiece(clientModel.GetDraggingElement(), moveEvent)
	}
}

func (clientModel *ClientModel) genMouseUp() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if clientModel.GetIsMouseDown() && len(i) > 0 {
			i[0].Call("preventDefault")
			clientModel.handleClickEnd(i[0])
		}
		return 0
	})
}

func (clientModel *ClientModel) genTouchEnd() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if clientModel.GetIsMouseDown() && len(i) > 0 {
			i[0].Call("preventDefault")
			touch := i[0].Get("changedTouches").Index(0)
			clientModel.handleClickEnd(touch)
		}
		return 0
	})
}

func (cm *ClientModel) handleClickEnd(event js.Value) {
	elDragging := cm.GetDraggingElement()
	cm.SetDraggingElement(js.Undefined())
	_, _, _, _, gridX, gridY := cm.getEventMousePosition(event)
	newPosition := cm.getPositionFromGrid(uint8(gridX), uint8(gridY))
	moveRequest := model.MoveRequest{cm.positionOriginal, model.Move{
		int8(newPosition.File) - int8(cm.positionOriginal.File),
		int8(newPosition.Rank) - int8(cm.positionOriginal.Rank),
	},
	}
	if cm.gameType == Local || cm.playerColor == cm.game.Turn() {
		go func() {
			cm.takeMove(moveRequest, newPosition, elDragging)
			elDragging.Get("style").Set("transform",
				cm.GetDraggingOriginalTransform())
			removeClass(elDragging, "dragging")
		}()
	} else {
		elDragging.Get("style").Set("transform",
			cm.GetDraggingOriginalTransform())
		removeClass(elDragging, "dragging")
	}
	cm.SetIsMouseDown(false)
}

func (cm *ClientModel) takeMove(
	moveRequest model.MoveRequest, newPos model.Position, elMoving js.Value) {
	successfulRemoteMove := false
	if cm.GetGameType() == Remote {
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
	err := cm.MakeMove(moveRequest)
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
	log.SetPrefix("listenForOpponentMove: ")
	// TODO need to write to the endRemoteGame chan when game is over
	retries := 0
	maxRetries := 10
	for true {
		select {
		case <-endRemoteGame:
			return
		default:
		}
		resp, err := cm.client.Get(cm.matchingServerURI + "sync")
		if err == nil {
			defer resp.Body.Close()
			opponentMove := model.MoveRequest{}
			json.NewDecoder(resp.Body).Decode(&opponentMove)
			err := cm.MakeMove(opponentMove)
			if err != nil {
				println("FATAL: We do not expect an invalid move from the opponent")
			}
			newPos := model.Position{
				opponentMove.Position.File + uint8(opponentMove.Move.X),
				opponentMove.Position.Rank + uint8(opponentMove.Move.Y),
			}
			originalPosClass :=
				getPositionClass(opponentMove.Position, cm.GetPlayerColor())
			elements := cm.document.Call("getElementsByClassName", originalPosClass)
			elMoving := elements.Index(0)
			cm.viewHandleMove(opponentMove, newPos, elMoving)
		} else {
			time.Sleep(500 * time.Millisecond)
			retries++
			if retries >= maxRetries {
				log.Printf("Reached maxRetries on uri=%s retries=%d",
					cm.matchingServerURI+"sync", maxRetries)
				return
			}
		}
	}
}

func (clientModel *ClientModel) genBeginMatchmaking() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if !clientModel.GetIsMatchmaking() && !clientModel.GetIsMatched() {
			go clientModel.lookForMatch()
		}
		return 0
	})
}

func (clientModel *ClientModel) lookForMatch() {
	clientModel.SetIsMatchmaking(true)
	buttonLoader := clientModel.buttonBeginLoading(
		clientModel.document.Call(
			"getElementById", "beginMatchmakingButton"))
	if !clientModel.GetHasSession() {
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
		clientModel.SetPlayerName(username)
		clientModel.SetHasSession(true)
	}
	resp, err := clientModel.client.Get(clientModel.matchingServerURI + "match")
	if err == nil {
		var matchResponse webserver.MatchedResponse
		json.NewDecoder(resp.Body).Decode(&matchResponse)
		resp.Body.Close()
		clientModel.SetPlayerColor(matchResponse.Color)
		clientModel.SetOpponentName(matchResponse.OpponentName)
		clientModel.resetGame()
		// - TODO once matched briefly display matched icon?
		// - TODO once matched set and display time remaining
		clientModel.SetGameType(Remote)
		clientModel.SetIsMatched(true)
		clientModel.SetIsMatchmaking(false)
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
	if cm.GetPlayerColor() == model.White {
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
	clientModel.SetGame(&game)
	clientModel.viewClearBoard()
	clientModel.viewInitBoard(clientModel.playerColor)
}
