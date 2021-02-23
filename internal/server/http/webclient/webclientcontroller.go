// +build webclient

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Ekotlikoff/gochess/internal/model"
	"github.com/Ekotlikoff/gochess/internal/server/http/webserver"
	"github.com/Ekotlikoff/gochess/internal/server/match"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

var debug bool = false
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
	js.Global().Set("resign", clientModel.genResign())
	js.Global().Set("draw", clientModel.genDraw())
	js.Global().Set("onclick", clientModel.genGlobalOnclick())
	clientModel.document.Call("getElementById",
		"gameover_modal_close").Set("onclick",
		clientModel.genCloseModalOnClick())
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

func (cm *ClientModel) genGlobalOnclick() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		gameoverModal := cm.document.Call("getElementById", "gameover_modal")
		if i[0].Get("target").Equal(gameoverModal) {
			cm.closeGameoverModal()
		}
		return 0
	})
}

func (cm *ClientModel) genCloseModalOnClick() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		cm.closeGameoverModal()
		return 0
	})
}

func (clientModel *ClientModel) closeGameoverModal() {
	removeClass(clientModel.document.Call("getElementById", "gameover_modal"),
		"gameover_modal")
	addClass(clientModel.document.Call("getElementById", "gameover_modal"),
		"hidden")
	clientModel.viewSetMatchMakingControls()
	clientModel.viewClearMatchDetails()
	clientModel.SetGameType(Local)
	clientModel.SetIsMatched(false)
	clientModel.ResetRemoteMatchModel()
	clientModel.resetGame()
}

func (clientModel *ClientModel) handleClickStart(
	this js.Value, event js.Value) {
	clientModel.LockMouseDown()
	clientModel.SetDraggingElement(this)
	positionOriginal, err := clientModel.getGamePositionFromPieceElement(this)
	if err != nil {
		log.Println("ERROR: Issue getting position from element,", err)
		return
	}
	clientModel.positionOriginal = positionOriginal
	clientModel.SetDraggingPiece(clientModel.positionOriginal)
	if clientModel.GetDraggingPiece() == nil {
		if debug {
			log.Println("ERROR: Clicked a piece that is not on the board")
			log.Println(clientModel.positionOriginal)
			log.Println(clientModel.GetBoardString())
		}
		clientModel.UnlockMouseDown()
		return
	}
	addClass(clientModel.GetDraggingElement(), "dragging")
	clientModel.SetDraggingOriginalTransform(
		clientModel.GetDraggingElement().Get("style").Get("transform"))
}

func (cm *ClientModel) getGamePositionFromPieceElement(
	piece js.Value) (model.Position, error) {
	className := piece.Get("className").String()
	classElements := strings.Split(className, " ")
	for i := range classElements {
		if strings.Contains(classElements[i], "square") {
			posString := strings.Split(classElements[i], "-")[1]
			x, err := strconv.Atoi(string(posString[0]))
			y, err := strconv.Atoi(string(posString[1]))
			if err != nil {
				return model.Position{}, err
			}
			x--
			y--
			if cm.GetPlayerColor() == model.Black {
				x = 7 - x
				y = 7 - y
			}
			return model.Position{uint8(x), uint8(y)}, nil
		}
	}
	return model.Position{},
		errors.New("Unable to convert class to position: " + className)
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
	pieceDragging := cm.GetDraggingPiece()
	var promoteTo *model.PieceType
	if pieceDragging.PieceType() == model.Pawn &&
		(newPosition.Rank == 0 || newPosition.Rank == 7) {
		promoteTo = cm.handlePromotion()
	}
	moveRequest := model.MoveRequest{cm.positionOriginal, model.Move{
		int8(newPosition.File) - int8(cm.positionOriginal.File),
		int8(newPosition.Rank) - int8(cm.positionOriginal.Rank)},
		promoteTo,
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
	cm.UnlockMouseDown()
}

func (cm *ClientModel) handlePromotion() *model.PieceType {
	// TODO Allow user selection of promoteTo
	promoteTo := model.Queen
	return &promoteTo
}

func (cm *ClientModel) takeMove(
	moveRequest model.MoveRequest, newPos model.Position, elMoving js.Value) {
	successfulRemoteMove := false
	if cm.GetGameType() == Remote {
		movePayloadBuf := new(bytes.Buffer)
		json.NewEncoder(movePayloadBuf).Encode(moveRequest)
		resp, err := cm.client.Post("sync", ctp, movePayloadBuf)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return
			}
			successfulRemoteMove = true
		}
	}
	err := cm.MakeMove(moveRequest)
	if err == nil {
		cm.ClearRequestedDraw()
		cm.viewHandleMove(moveRequest, newPos, elMoving)
	} else {
		if debug {
			log.Println(err)
		}
		if successfulRemoteMove {
			// TODO handle strange case where http call was successful but local
			// game did not accept the move.
			log.Println("FATAL: We do not expect an unsuccessful local move when remote succeeds")
		}
	}
}

func (cm *ClientModel) listenForSyncUpdate() {
	log.SetPrefix("listenForSyncUpdate: ")
	endRemoteGame := cm.remoteMatchModel.endRemoteGameChan
	retries := 0
	maxRetries := 5
	for true {
		select {
		case <-endRemoteGame:
			return
		default:
		}
		resp, err := cm.client.Get("sync")
		if err == nil {
			defer resp.Body.Close()
			retries = 0
			if resp.StatusCode == 200 {
				opponentMove := model.MoveRequest{}
				json.NewDecoder(resp.Body).Decode(&opponentMove)
				err := cm.MakeMove(opponentMove)
				if err != nil {
					log.Println("FATAL: We do not expect an invalid move from the opponent.")
				}
				cm.ClearRequestedDraw()
				newPos := model.Position{
					opponentMove.Position.File + uint8(opponentMove.Move.X),
					opponentMove.Position.Rank + uint8(opponentMove.Move.Y),
				}
				originalPosClass :=
					getPositionClass(opponentMove.Position, cm.GetPlayerColor())
				elements :=
					cm.document.Call("getElementsByClassName", originalPosClass)
				elMoving := elements.Index(0)
				cm.viewHandleMove(opponentMove, newPos, elMoving)
			}
		} else {
			log.Println(err)
			time.Sleep(500 * time.Millisecond)
			retries++
			if retries >= maxRetries {
				log.Printf("Reached maxRetries on uri=%s retries=%d",
					"sync", maxRetries)
				close(cm.remoteMatchModel.endRemoteGameChan)
				return
			}
		}
	}
}

func (cm *ClientModel) listenForAsyncUpdate() {
	log.SetPrefix("listenForAsyncUpdate: ")
	endRemoteGame := cm.remoteMatchModel.endRemoteGameChan
	retries := 0
	maxRetries := 5
	for true {
		select {
		case <-endRemoteGame:
			return
		default:
		}
		resp, err := cm.client.Get("async")
		if err == nil {
			defer resp.Body.Close()
			retries = 0
			if resp.StatusCode == 200 {
				asyncResponse := matchserver.ResponseAsync{}
				json.NewDecoder(resp.Body).Decode(&asyncResponse)
				if asyncResponse.GameOver {
					close(cm.remoteMatchModel.endRemoteGameChan)
					winType := ""
					if asyncResponse.Resignation {
						winType = "resignation"
					} else if asyncResponse.Draw {
						winType = "draw"
						cm.ClearRequestedDraw()
					} else if asyncResponse.Timeout {
						winType = "timeout"
					} else {
						winType = "mate"
					}
					log.Println("Winner:", asyncResponse.Winner, "by", winType)
					cm.viewSetGameOver(asyncResponse.Winner, winType)
					return
				} else if asyncResponse.RequestToDraw {
					log.Println("Requested draw")
					cm.SetRequestedDraw(cm.GetOpponentColor(),
						!cm.GetRequestedDraw(cm.GetOpponentColor()))
				}
			}
		} else {
			log.Println(err)
			time.Sleep(500 * time.Millisecond)
			retries++
			if retries >= maxRetries {
				log.Printf("Reached maxRetries on uri=%s retries=%d",
					"async", maxRetries)
				close(cm.remoteMatchModel.endRemoteGameChan)
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

func (clientModel *ClientModel) genResign() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		requestBuf := new(bytes.Buffer)
		request := matchserver.RequestAsync{Resign: true}
		json.NewEncoder(requestBuf).Encode(request)
		go clientModel.client.Post("async", ctp, requestBuf)
		return 0
	})
}

func (clientModel *ClientModel) genDraw() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		go clientModel.sendDraw()
		return 0
	})
}

func (clientModel *ClientModel) sendDraw() {
	requestBuf := new(bytes.Buffer)
	request := matchserver.RequestAsync{RequestToDraw: true}
	json.NewEncoder(requestBuf).Encode(request)
	_, err := clientModel.client.Post("async", ctp, requestBuf)
	if err == nil {
		clientModel.SetRequestedDraw(clientModel.GetPlayerColor(),
			!clientModel.GetRequestedDraw(clientModel.GetPlayerColor()))
	}
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
		resp, err := clientModel.client.Post("session", ctp, credentialsBuf)
		if err == nil {
			resp.Body.Close()
		}
		clientModel.SetPlayerName(username)
		clientModel.SetHasSession(true)
	}
	retries := 0
	maxRetries := 6000
	var resp *http.Response
	var err error
	for true {
		resp, err = clientModel.client.Get("match")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
		retries++
		if retries >= maxRetries {
			log.Printf("Reached maxRetries on uri=%s retries=%d",
				"match", maxRetries)
		}
	}
	var matchResponse webserver.MatchedResponse
	json.NewDecoder(resp.Body).Decode(&matchResponse)
	clientModel.SetPlayerColor(matchResponse.Color)
	clientModel.SetOpponentName(matchResponse.OpponentName)
	clientModel.SetMaxTimeMs(matchResponse.MaxTimeMs)
	clientModel.resetGame()
	// - TODO once matched briefly display matched icon?
	clientModel.SetGameType(Remote)
	clientModel.SetIsMatched(true)
	clientModel.SetIsMatchmaking(false)
	buttonLoader.Call("remove")
	clientModel.remoteMatchModel.endRemoteGameChan = make(chan bool, 0)
	clientModel.viewSetMatchControls()
	go clientModel.matchDetailsUpdateLoop()
	go clientModel.listenForSyncUpdate()
	go clientModel.listenForAsyncUpdate()
}

func (clientModel *ClientModel) matchDetailsUpdateLoop() {
	for true {
		clientModel.viewSetMatchDetails()
		time.Sleep(1000 * time.Millisecond)
		select {
		case <-clientModel.remoteMatchModel.endRemoteGameChan:
			return
		default:
		}
		turn := clientModel.game.Turn()
		clientModel.AddPlayerElapsedMs(turn, 1000)
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
