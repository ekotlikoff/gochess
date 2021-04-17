// +build webclient

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"
	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
)

var (
	debug bool   = true
	ctp   string = "application/json"
)

func (cm *ClientModel) initController() {
	cm.document.Call("addEventListener", "mousemove", cm.genMouseMove(), false)
	cm.document.Call("addEventListener", "touchmove", cm.genTouchMove(), false)
	cm.document.Call("addEventListener", "mouseup", cm.genMouseUp(), false)
	cm.document.Call("addEventListener", "touchend", cm.genTouchEnd(), false)
	cm.board.Call("addEventListener", "contextmenu",
		js.FuncOf(preventDefault), false)
	js.Global().Set("beginMatchmaking", cm.genBeginMatchmaking())
	js.Global().Set("resign", cm.genResign())
	js.Global().Set("draw", cm.genDraw())
	js.Global().Set("onclick", cm.genGlobalOnclick())
	cm.document.Call("getElementById", "gameover_modal_close").Set("onclick",
		cm.genCloseModalOnClick())
}

func (cm *ClientModel) genMouseDown() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if len(i) > 0 && !cm.GetIsMouseDown() {
			i[0].Call("preventDefault")
			cm.handleClickStart(this, i[0])
		}
		return 0
	})
}

func (cm *ClientModel) genTouchStart() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if len(i) > 0 && !cm.GetIsMouseDown() {
			i[0].Call("preventDefault")
			touch := i[0].Get("touches").Index(0)
			cm.handleClickStart(this, touch)
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

func (cm *ClientModel) closeGameoverModal() {
	removeClass(cm.document.Call("getElementById", "gameover_modal"),
		"gameover_modal")
	addClass(cm.document.Call("getElementById", "gameover_modal"), "hidden")
	cm.viewSetMatchMakingControls()
	cm.viewClearMatchDetails()
	cm.SetGameType(Local)
	cm.SetIsMatched(false)
	cm.ResetRemoteMatchModel()
	cm.resetGame()
}

func (cm *ClientModel) handleClickStart(
	this js.Value, event js.Value) {
	cm.LockMouseDown()
	cm.SetDraggingElement(this)
	positionOriginal, err := cm.getGamePositionFromPieceElement(this)
	if err != nil {
		log.Println("ERROR: Issue getting position from element,", err)
		return
	}
	cm.positionOriginal = positionOriginal
	cm.SetDraggingPiece(cm.positionOriginal)
	if cm.GetDraggingPiece() == nil {
		if debug {
			log.Println("ERROR: Clicked a piece that is not on the board")
			log.Println(cm.positionOriginal)
			log.Println(cm.GetBoardString())
		}
		cm.UnlockMouseDown()
		return
	}
	addClass(cm.GetDraggingElement(), "dragging")
	cm.SetDraggingOriginalTransform(
		cm.GetDraggingElement().Get("style").Get("transform"))
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

func (cm *ClientModel) genMouseMove() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		i[0].Call("preventDefault")
		cm.handleMoveEvent(i[0])
		return 0
	})
}

func (cm *ClientModel) genTouchMove() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		i[0].Call("preventDefault")
		touch := i[0].Get("touches").Index(0)
		cm.handleMoveEvent(touch)
		return 0
	})
}

func (cm *ClientModel) handleMoveEvent(moveEvent js.Value) {
	if cm.GetIsMouseDown() {
		cm.viewDragPiece(cm.GetDraggingElement(), moveEvent)
	}
}

func (cm *ClientModel) genMouseUp() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if cm.GetIsMouseDown() && len(i) > 0 {
			i[0].Call("preventDefault")
			cm.handleClickEnd(i[0])
		}
		return 0
	})
}

func (cm *ClientModel) genTouchEnd() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if cm.GetIsMouseDown() && len(i) > 0 {
			i[0].Call("preventDefault")
			touch := i[0].Get("changedTouches").Index(0)
			cm.handleClickEnd(touch)
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
	err := cm.MakeMove(moveRequest)
	if err == nil {
		cm.ClearRequestedDraw()
		cm.viewHandleMove(moveRequest, newPos, elMoving)
	} else {
		if debug {
			log.Println(err)
		}
		return
	}
	if cm.GetGameType() == Remote {
		if cm.backendType == HttpBackend {
			movePayloadBuf := new(bytes.Buffer)
			json.NewEncoder(movePayloadBuf).Encode(moveRequest)
			resp, err := retryWrapper(
				func() (*http.Response, error) {
					return cm.client.Post("http/sync", ctp, movePayloadBuf)
				},
				"POST http/sync", 0,
				func() { cm.remoteGameEnd() },
			)
			if err != nil {
				log.Println("FATAL: Failed to send move to server: ", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				// TODO handle strange case where local accepts the
				// move but remote does not.
				log.Println("FATAL: We do not expect a remote rejection when local move succeeds")
				return
			}
		} else if cm.backendType == WebsocketBackend {
			message := matchserver.WebsocketRequest{
				WebsocketRequestType: matchserver.RequestSyncT,
				RequestSync:          moveRequest,
			}
			jsonMsg, _ := json.Marshal(message)
			cm.GetWSConn().Call("send", string(jsonMsg))
		}
	}
}

func (cm *ClientModel) listenForSyncUpdate() {
	log.SetPrefix("listenForSyncUpdate: ")
	if cm.backendType == HttpBackend {
		cm.listenForSyncUpdateHttp()
	}
}

func (cm *ClientModel) listenForSyncUpdateHttp() {
	for true {
		log.SetPrefix("listenForSyncUpdate: ")
		select {
		case <-cm.remoteMatchModel.endRemoteGameChan:
			return
		default:
		}
		resp, err := retryWrapper(
			func() (*http.Response, error) {
				return cm.client.Get("http/sync")
			},
			"GET http/sync", 0,
			func() { cm.remoteGameEnd() },
		)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			opponentMove := model.MoveRequest{}
			json.NewDecoder(resp.Body).Decode(&opponentMove)
			cm.handleSyncUpdate(opponentMove)
		}
	}
}

func (cm *ClientModel) handleSyncUpdate(opponentMove model.MoveRequest) {
	err := cm.MakeMove(opponentMove)
	if err != nil {
		log.Println("FATAL: We do not expect an invalid move from the opponent.")
		return
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

func (cm *ClientModel) listenForAsyncUpdate() {
	log.SetPrefix("listenForAsyncUpdate: ")
	if cm.backendType == HttpBackend {
		cm.listenForAsyncUpdateHttp()
	}
}

func (cm *ClientModel) listenForAsyncUpdateHttp() {
	for true {
		log.SetPrefix("listenForAsyncUpdate: ")
		select {
		case <-cm.remoteMatchModel.endRemoteGameChan:
			return
		default:
		}
		resp, err := retryWrapper(
			func() (*http.Response, error) {
				return cm.client.Get("http/async")
			},
			"http/async", 0,
			func() { cm.remoteGameEnd() },
		)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			asyncResponse := matchserver.ResponseAsync{}
			json.NewDecoder(resp.Body).Decode(&asyncResponse)
			cm.handleResponseAsync(asyncResponse)
		}
	}
}

func (cm *ClientModel) handleResponseAsync(responseAsync matchserver.ResponseAsync) {
	if responseAsync.GameOver {
		cm.remoteGameEnd()
		winType := ""
		if responseAsync.Resignation {
			winType = "resignation"
		} else if responseAsync.Draw {
			winType = "draw"
			cm.ClearRequestedDraw()
		} else if responseAsync.Timeout {
			winType = "timeout"
		} else {
			winType = "mate"
		}
		log.Println("Winner:", responseAsync.Winner, "by", winType)
		cm.viewSetGameOver(responseAsync.Winner, winType)
		return
	} else if responseAsync.RequestToDraw {
		log.Println("Requested draw")
		cm.SetRequestedDraw(cm.GetOpponentColor(),
			!cm.GetRequestedDraw(cm.GetOpponentColor()))
	}
}

func (cm *ClientModel) genBeginMatchmaking() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		if !cm.GetIsMatchmaking() && !cm.GetIsMatched() {
			go cm.lookForMatch()
		}
		return 0
	})
}

func (cm *ClientModel) genResign() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		requestBuf := new(bytes.Buffer)
		request := matchserver.RequestAsync{Resign: true}
		json.NewEncoder(requestBuf).Encode(request)
		if cm.backendType == HttpBackend {
			go cm.client.Post("http/async", ctp, requestBuf)
		} else if cm.backendType == WebsocketBackend {
			message := matchserver.WebsocketRequest{
				WebsocketRequestType: matchserver.RequestAsyncT,
				RequestAsync:         request,
			}
			jsonMsg, _ := json.Marshal(message)
			cm.GetWSConn().Call("send", string(jsonMsg))
		}
		return 0
	})
}

func (cm *ClientModel) genDraw() js.Func {
	return js.FuncOf(func(this js.Value, i []js.Value) interface{} {
		go cm.sendDraw()
		return 0
	})
}

func (cm *ClientModel) sendDraw() {
	requestBuf := new(bytes.Buffer)
	request := matchserver.RequestAsync{RequestToDraw: true}
	json.NewEncoder(requestBuf).Encode(request)
	if cm.backendType == HttpBackend {
		_, err := cm.client.Post("http/async", ctp, requestBuf)
		if err != nil {
			return
		}
	} else if cm.backendType == WebsocketBackend {
		message := matchserver.WebsocketRequest{
			WebsocketRequestType: matchserver.RequestAsyncT,
			RequestAsync:         request,
		}
		jsonMsg, _ := json.Marshal(message)
		cm.GetWSConn().Call("send", string(jsonMsg))
	}
	cm.SetRequestedDraw(cm.GetPlayerColor(),
		!cm.GetRequestedDraw(cm.GetPlayerColor()))
}

func (cm *ClientModel) lookForMatch() {
	cm.SetIsMatchmaking(true)
	buttonLoader := cm.buttonBeginLoading(
		cm.document.Call("getElementById", "beginMatchmakingButton"))
	if !cm.GetHasSession() {
		username := cm.document.Call(
			"getElementById", "username").Get("value").String()
		credentialsBuf := new(bytes.Buffer)
		credentials := gateway.Credentials{username}
		json.NewEncoder(credentialsBuf).Encode(credentials)
		resp, err := cm.client.Post("session", ctp, credentialsBuf)
		if err == nil {
			resp.Body.Close()
		}
		cm.SetPlayerName(username)
		cm.SetHasSession(true)
	}
	var matchResponse matchserver.MatchedResponse
	var err error
	if cm.backendType == HttpBackend {
		matchResponse, err = cm.httpMatch(buttonLoader)
	} else if cm.backendType == WebsocketBackend {
		matchResponse, err = cm.wsMatch(buttonLoader)
	}
	if err != nil {
		return
	}
	cm.SetPlayerColor(matchResponse.Color)
	cm.SetOpponentName(matchResponse.OpponentName)
	cm.SetMaxTimeMs(matchResponse.MaxTimeMs)
	cm.resetGame()
	// - TODO once matched briefly display matched icon?
	cm.SetGameType(Remote)
	cm.SetIsMatched(true)
	cm.SetIsMatchmaking(false)
	buttonLoader.Call("remove")
	cm.remoteMatchModel.endRemoteGameChan = make(chan bool, 0)
	cm.viewSetMatchControls()
	go cm.matchDetailsUpdateLoop()
	go cm.listenForSyncUpdate()
	go cm.listenForAsyncUpdate()
}

func (cm *ClientModel) httpMatch(buttonLoader js.Value,
) (matchserver.MatchedResponse, error) {
	resp, err := retryWrapper(
		func() (*http.Response, error) {
			return cm.client.Get("http/match")
		},
		"http/match", 200,
		func() {
			cm.SetIsMatchmaking(false)
			buttonLoader.Call("remove")
		},
	)
	var matchedResponse matchserver.MatchedResponse
	if err != nil {
		return matchedResponse, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&matchedResponse)
	return matchedResponse, nil
}

func (cm *ClientModel) wsMatch(buttonLoader js.Value,
) (matchserver.MatchedResponse, error) {
	u := "ws://" + cm.origin + "/ws"
	ws := js.Global().Get("WebSocket").New(u)
	retries := 0
	maxRetries := 100
	matchedChan := make(chan matchserver.MatchedResponse)
	var matchResponse matchserver.MatchedResponse
	for true {
		if ws.Get("readyState").Equal(js.Global().Get("WebSocket").Get("OPEN")) {
			cm.SetWSConn(ws)
			if debug {
				log.Println("Websocket connection successfully initiated")
			}
			message := matchserver.WebsocketRequest{
				WebsocketRequestType: matchserver.RequestAsyncT,
				RequestAsync:         matchserver.RequestAsync{Match: true},
			}
			jsonMsg, _ := json.Marshal(message)
			ws.Call("send", string(jsonMsg))
			go cm.wsListener(matchedChan)
			break
		}
		time.Sleep(100 * time.Millisecond)
		retries++
		if retries > maxRetries {
			cm.SetIsMatchmaking(false)
			buttonLoader.Call("remove")
			log.Println("ERROR: Error opening websocket connection")
			return matchResponse, errors.New("Error opening websocket connection")
		}
	}
	return <-matchedChan, nil
}

func (cm *ClientModel) wsListener(matchedChan chan matchserver.MatchedResponse) {
	ws := cm.GetWSConn()
	ws.Set("onmessage",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			jsonString := args[0].Get("data").String()
			log.Println(jsonString)
			message := matchserver.WebsocketResponse{}
			json.Unmarshal([]byte(jsonString), &message)
			switch message.WebsocketResponseType {
			case matchserver.MatchStartT:
				matchedChan <- message.MatchedResponse
			case matchserver.OpponentPlayedMoveT:
				cm.handleSyncUpdate(message.OpponentPlayedMove)
			case matchserver.ResponseAsyncT:
				cm.handleResponseAsync(message.ResponseAsync)
			}
			return nil
		}))
}

func (cm *ClientModel) matchDetailsUpdateLoop() {
	for true {
		cm.viewSetMatchDetails()
		time.Sleep(1000 * time.Millisecond)
		select {
		case <-cm.remoteMatchModel.endRemoteGameChan:
			return
		default:
		}
		turn := cm.game.Turn()
		cm.AddPlayerElapsedMs(turn, 1000)
	}
}

func (cm *ClientModel) getEventMousePosition(event js.Value) (
	int, int, int, int, int, int) {
	rect := cm.board.Call("getBoundingClientRect")
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

func (cm *ClientModel) resetGame() {
	game := model.NewGame()
	cm.SetGame(game)
	cm.viewClearBoard()
	cm.viewInitBoard(cm.playerColor)
}

func retryWrapper(f func() (*http.Response, error), uri string, successCode int,
	onMaxRetries func()) (*http.Response, error) {
	retries := 0
	maxRetries := 5
	for true {
		resp, err := f()
		if err != nil || (resp.StatusCode != successCode && successCode != 0) {
			log.Println(err)
			time.Sleep(500 * time.Millisecond)
			retries++
			if retries >= maxRetries {
				log.Printf("Reached maxRetries on uri=%s retries=%d",
					uri, maxRetries)
				onMaxRetries()
				return nil, err
			}
		} else {
			return resp, nil
		}
	}
	return nil, nil
}

func (cm *ClientModel) remoteGameEnd() {
	select {
	case <-cm.remoteMatchModel.endRemoteGameChan:
		return
	default:
		close(cm.remoteMatchModel.endRemoteGameChan)
		cm.viewSetGameOver("", "error")
	}
}
