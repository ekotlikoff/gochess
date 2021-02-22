// +build webclient

package main

import (
	"github.com/Ekotlikoff/gochess/internal/model"
	"net/http"
	"sync"
	"syscall/js"
)

type GameType uint8

const (
	Local  = GameType(iota)
	Remote = GameType(iota)
	quiet  = false
)

type ClientModel struct {
	cmMutex                  sync.RWMutex
	gameType                 GameType
	playerColor              model.Color
	elDragging               js.Value
	draggingOrigTransform    js.Value
	isMouseDown              bool
	positionOriginal         model.Position
	isMatchmaking, isMatched bool
	playerName               string
	hasSession               bool
	gameMutex                sync.RWMutex
	game                     *model.Game
	remoteMatchModel         RemoteMatchModel
	// Unchanging elements
	document          js.Value
	board             js.Value
	matchingServerURI string
	client            *http.Client
}

type RemoteMatchModel struct {
	opponentName          string
	maxTimeMs             int64
	playerElapsedMs       int64
	opponentElapsedMs     int64
	opponentRequestedDraw bool
	playerRequestedDraw   bool
	endRemoteGameChan     chan bool
}

func (cm *ClientModel) ResetRemoteMatchModel() {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.remoteMatchModel = RemoteMatchModel{}
}

func (cm *ClientModel) GetGameType() GameType {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.gameType
}

func (cm *ClientModel) SetGameType(gameType GameType) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.gameType = gameType
}

func (cm *ClientModel) SetGame(game *model.Game) {
	cm.gameMutex.Lock()
	defer cm.gameMutex.Unlock()
	cm.game = game
}

func (cm *ClientModel) GetTurn() model.Color {
	cm.gameMutex.Lock()
	defer cm.gameMutex.Unlock()
	return cm.game.Turn()
}

func (cm *ClientModel) GetPointAdvantage(color model.Color) int8 {
	cm.gameMutex.Lock()
	defer cm.gameMutex.Unlock()
	return cm.game.PointAdvantage(color)
}

func (cm *ClientModel) MakeMove(moveRequest model.MoveRequest) error {
	cm.gameMutex.Lock()
	defer cm.gameMutex.Unlock()
	return cm.game.Move(moveRequest)
}

func (cm *ClientModel) GetPlayerColor() model.Color {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.playerColor
}

func (cm *ClientModel) GetOpponentColor() model.Color {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	opponentColor := model.Black
	if cm.playerColor == opponentColor {
		opponentColor = model.White
	}
	return opponentColor
}

func (cm *ClientModel) SetPlayerColor(color model.Color) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.playerColor = color
}

func (cm *ClientModel) GetDraggingElement() js.Value {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.elDragging
}

func (cm *ClientModel) SetDraggingElement(el js.Value) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.elDragging = el
}

func (cm *ClientModel) GetDraggingOriginalTransform() js.Value {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.draggingOrigTransform
}

func (cm *ClientModel) SetDraggingOriginalTransform(el js.Value) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.draggingOrigTransform = el
}

func (cm *ClientModel) GetIsMouseDown() bool {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.isMouseDown
}

func (cm *ClientModel) SetIsMouseDown(isMouseDown bool) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.isMouseDown = isMouseDown
}

func (cm *ClientModel) GetClickOriginalPosition() model.Position {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.positionOriginal
}

func (cm *ClientModel) SetClickOriginalPosition(position model.Position) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.positionOriginal = position
}

func (cm *ClientModel) GetIsMatchmaking() bool {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.isMatchmaking
}

func (cm *ClientModel) SetIsMatchmaking(isMatchmaking bool) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.isMatchmaking = isMatchmaking
}

func (cm *ClientModel) GetIsMatched() bool {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.isMatched
}

func (cm *ClientModel) SetIsMatched(isMatched bool) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.isMatched = isMatched
}

func (cm *ClientModel) GetPlayerName() string {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.playerName
}

func (cm *ClientModel) SetPlayerName(name string) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.playerName = name
}

func (cm *ClientModel) GetOpponentName() string {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.remoteMatchModel.opponentName
}

func (cm *ClientModel) SetOpponentName(name string) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.remoteMatchModel.opponentName = name
}

func (cm *ClientModel) GetMaxTimeMs() int64 {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.remoteMatchModel.maxTimeMs
}

func (cm *ClientModel) SetMaxTimeMs(maxTimeMs int64) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.remoteMatchModel.maxTimeMs = maxTimeMs
}

func (cm *ClientModel) GetPlayerElapsedMs(color model.Color) int64 {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	if color == cm.playerColor {
		return cm.remoteMatchModel.playerElapsedMs
	}
	return cm.remoteMatchModel.opponentElapsedMs
}

func (cm *ClientModel) AddPlayerElapsedMs(color model.Color, elapsedMs int64) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	if color == cm.playerColor {
		cm.remoteMatchModel.playerElapsedMs += elapsedMs
	} else {
		cm.remoteMatchModel.opponentElapsedMs += elapsedMs
	}
}

func (cm *ClientModel) GetRequestedDraw(color model.Color) bool {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	if color == cm.playerColor {
		return cm.remoteMatchModel.playerRequestedDraw
	} else {
		return cm.remoteMatchModel.opponentRequestedDraw
	}
}

func (cm *ClientModel) ClearRequestedDraw() {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.remoteMatchModel.playerRequestedDraw = false
	cm.remoteMatchModel.opponentRequestedDraw = false
}

func (cm *ClientModel) SetRequestedDraw(color model.Color, requestedDraw bool) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	if color == cm.playerColor {
		cm.remoteMatchModel.playerRequestedDraw = requestedDraw
	} else {
		cm.remoteMatchModel.opponentRequestedDraw = requestedDraw
	}
}

func (cm *ClientModel) GetHasSession() bool {
	cm.cmMutex.RLock()
	defer cm.cmMutex.RUnlock()
	return cm.hasSession
}

func (cm *ClientModel) SetHasSession(hasSession bool) {
	cm.cmMutex.Lock()
	defer cm.cmMutex.Unlock()
	cm.hasSession = hasSession
}
