// +build webclient

package main

import (
	"github.com/Ekotlikoff/gochess/internal/model"
	"net/http"
	"net/http/cookiejar"
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
	gameType                 GameType
	game                     *model.Game
	playerColor              model.Color
	elDragging               js.Value
	draggingOrigTransform    js.Value
	isMouseDown              bool
	positionOriginal         model.Position
	isMatchmaking, isMatched bool
	playerName               string
	opponentName             string
	hasSession               bool
	mutex                    sync.RWMutex
	gameMutex                sync.RWMutex
	// Unchanging elements
	document          js.Value
	board             js.Value
	matchingServerURI string
	client            *http.Client
	endRemoteGameChan chan bool
}

func (cm *ClientModel) GetGameType() GameType {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.gameType
}

func (cm *ClientModel) GetTurn() model.Color {
	cm.gameMutex.Lock()
	defer cm.gameMutex.Unlock()
	return cm.game.Turn()
}

func (cm *ClientModel) MakeMove(moveRequest model.MoveRequest) error {
	cm.gameMutex.Lock()
	defer cm.gameMutex.Unlock()
	return cm.game.Move(moveRequest)
}

func (cm *ClientModel) GetPlayerColor() model.Color {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.playerColor
}

func (cm *ClientModel) SetPlayerColor(color model.Color) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	cm.playerColor = color
}

func (cm *ClientModel) GetDraggingElement() js.Value {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.elDragging
}

func (cm *ClientModel) SetDraggingElement(el js.Value) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.elDragging = el
}

func (cm *ClientModel) GetIsMouseDown() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.isMouseDown
}

func (cm *ClientModel) SetIsMouseDown(isMouseDown bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.isMouseDown = isMouseDown
}

func (cm *ClientModel) GetClickOriginalPosition() model.Position {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.positionOriginal
}

func (cm *ClientModel) SetClickOriginalPosition(position model.Position) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.positionOriginal = position
}

func (cm *ClientModel) GetIsMatchmaking() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.isMatchmaking
}

func (cm *ClientModel) SetIsMatchmaking(isMatchmaking bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.isMatchmaking = isMatchmaking
}

func (cm *ClientModel) GetIsMatched() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.isMatched
}

func (cm *ClientModel) SetIsMatched(isMatched bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.isMatched = isMatched
}

func (cm *ClientModel) GetPlayerName() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.playerName
}

func (cm *ClientModel) SetPlayerName(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.playerName = name
}

func (cm *ClientModel) GetOpponentName() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.opponentName
}

func (cm *ClientModel) SetOpponentName(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.opponentName = name
}

func (cm *ClientModel) GetHasSession() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.hasSession
}

func (cm *ClientModel) SetHasSession(hasSession bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.hasSession = hasSession
}

func main() {
	done := make(chan struct{}, 0)
	game := model.NewGame()
	jar, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	clientModel := ClientModel{
		game: &game, playerColor: model.White,
		document: js.Global().Get("document"),
		board: js.Global().Get("document").Call(
			"getElementById", "board-layout-chessboard"),
		matchingServerURI: "http://192.168.1.166:8000/",
		client:            client,
	}
	clientModel.initController(quiet)
	clientModel.initStyle()
	clientModel.viewInitBoard(clientModel.playerColor)
	<-done
}
