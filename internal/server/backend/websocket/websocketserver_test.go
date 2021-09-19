package websocketserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"
	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
	"github.com/gorilla/websocket"
)

var (
	serverSession      *httptest.Server
	serverMatchAndPlay *httptest.Server
)

func init() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	close(exitChan)
	matchingServer.StartMatchServers(10, exitChan)
	serverSession = httptest.NewServer(http.HandlerFunc(gateway.Session))
	serverMatchAndPlay = httptest.NewServer(http.Handler(makeWebsocketHandler(&matchingServer)))
}

func TestWSMatch(t *testing.T) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	u := "ws" + strings.TrimPrefix(serverMatchAndPlay.URL, "http")
	wsDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}
	wsDialer2 := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client2.Jar,
	}
	ws, _, err := wsDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws2, _, err := wsDialer2.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws2.Close()
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         matchserver.RequestAsync{Match: true},
	}
	ws.WriteJSON(&message)
	ws2.WriteJSON(&message)
	wsResponse := matchserver.WebsocketResponse{}
	wsResponse2 := matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.WebsocketResponseType != matchserver.ResponseAsyncT ||
		wsResponse.ResponseAsync.MatchDetails.Color == wsResponse2.ResponseAsync.MatchDetails.Color {
		t.Error("Expected the two players to have different colors")
	}
	black := ws
	white := ws2
	whiteName := "player2"
	if wsResponse.ResponseAsync.MatchDetails.Color == model.White {
		black = ws2
		white = ws
		whiteName = "player1"
	}
	playerResp, enemyResp := makeMove(0, 0, 0, 0, white, black)
	if playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected invalid move response")
	}
	playerResp, enemyResp = makeMove(4, 1, 0, 2, white, black)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	if enemyResp.OpponentPlayedMove.Move.Y != 2 {
		t.Error("Expected opponent's move")
	}
	playerResp, enemyResp = makeMove(0, 6, 0, -1, black, white)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	playerResp, enemyResp = makeMove(3, 0, 4, 4, white, black)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	if enemyResp.OpponentPlayedMove.Move.Y != 4 {
		t.Error("Expected opponent's move")
	}
	playerResp, enemyResp = makeMove(0, 5, 0, -1, black, white)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	playerResp, enemyResp = makeMove(5, 0, -3, 3, white, black)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	playerResp, enemyResp = makeMove(0, 4, 0, -1, black, white)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	playerResp, enemyResp = makeMove(7, 4, -2, 2, white, black)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	if enemyResp.OpponentPlayedMove.Move.Y != 2 {
		t.Error("Expected opponent's move, got " +
			fmt.Sprint(enemyResp.OpponentPlayedMove.Move.Y))
	}
	black.ReadJSON(&playerResp)
	white.ReadJSON(&enemyResp)
	if !playerResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	if !enemyResp.ResponseAsync.GameOver ||
		enemyResp.ResponseAsync.Winner != whiteName {
		t.Error("Expected gameover and proper winner")
	}
	white.Close()
	black.Close()
}

func TestWSDraw(t *testing.T) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	u := "ws" + strings.TrimPrefix(serverMatchAndPlay.URL, "http")
	wsDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}
	wsDialer2 := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client2.Jar,
	}
	ws, _, err := wsDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws2, _, err := wsDialer2.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws2.Close()
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         matchserver.RequestAsync{Match: true},
	}
	ws.WriteJSON(&message)
	ws2.WriteJSON(&message)
	wsResponse := matchserver.WebsocketResponse{}
	wsResponse2 := matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.WebsocketResponseType != matchserver.ResponseAsyncT ||
		wsResponse.ResponseAsync.MatchDetails.Color == wsResponse2.ResponseAsync.MatchDetails.Color {
		t.Error("Expected the two players to have different colors")
	}
	black := ws
	white := ws2
	if wsResponse.ResponseAsync.MatchDetails.Color == model.White {
		black = ws2
		white = ws
	}
	playerResp, enemyResp := makeMove(0, 0, 0, 0, white, black)
	if playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected invalid move response")
	}
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	// Toggle off the request to draw
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	// Toggle back on
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	// Move toggles it off
	playerResp, enemyResp = makeMove(4, 1, 0, 2, white, black)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	// Finally draw
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, black, white)
	if !enemyResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	if !playerResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	white.Close()
	black.Close()
}

func TestWSResign(t *testing.T) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	u := "ws" + strings.TrimPrefix(serverMatchAndPlay.URL, "http")
	wsDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}
	wsDialer2 := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client2.Jar,
	}
	ws, _, err := wsDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws2, _, err := wsDialer2.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws2.Close()
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         matchserver.RequestAsync{Match: true},
	}
	ws.WriteJSON(&message)
	ws2.WriteJSON(&message)
	wsResponse := matchserver.WebsocketResponse{}
	wsResponse2 := matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.WebsocketResponseType != matchserver.ResponseAsyncT ||
		wsResponse.ResponseAsync.MatchDetails.Color == wsResponse2.ResponseAsync.MatchDetails.Color {
		t.Error("Expected the two players to have different colors")
	}
	black := ws
	white := ws2
	if wsResponse.ResponseAsync.MatchDetails.Color == model.White {
		black = ws2
		white = ws
	}
	playerResp, enemyResp := makeMove(0, 0, 0, 0, white, black)
	if playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected invalid move response")
	}
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	// Toggle off the request to draw
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	// Toggle back on
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, white, black)
	if !enemyResp.ResponseAsync.RequestToDraw {
		t.Error("Expected request to draw")
	}
	// Move toggles it off
	playerResp, enemyResp = makeMove(4, 1, 0, 2, white, black)
	if !playerResp.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{RequestToDraw: true}, black, white)
	// Finally resign
	playerResp, enemyResp = makeAsyncReq(
		matchserver.RequestAsync{Resign: true}, black, white)
	if !enemyResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	if !playerResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	white.Close()
	black.Close()
}

func TestWSRematch(t *testing.T) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	u := "ws" + strings.TrimPrefix(serverMatchAndPlay.URL, "http")
	wsDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}
	wsDialer2 := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client2.Jar,
	}
	ws, _, err := wsDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws2, _, err := wsDialer2.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws2.Close()
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         matchserver.RequestAsync{Match: true},
	}
	ws.WriteJSON(&message)
	ws2.WriteJSON(&message)
	wsResponse := matchserver.WebsocketResponse{}
	wsResponse2 := matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.WebsocketResponseType != matchserver.ResponseAsyncT ||
		wsResponse.ResponseAsync.MatchDetails.Color ==
			wsResponse2.ResponseAsync.MatchDetails.Color {
		t.Error("Expected the two players to have different colors")
	}
	black := ws
	white := ws2
	if wsResponse.ResponseAsync.MatchDetails.Color == model.White {
		black = ws2
		white = ws
	}
	playerResp, enemyResp := makeAsyncReq(
		matchserver.RequestAsync{Resign: true}, black, white)
	if !enemyResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	if !playerResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	message = matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         matchserver.RequestAsync{Match: true},
	}
	ws.WriteJSON(&message)
	ws2.WriteJSON(&message)
	wsResponse = matchserver.WebsocketResponse{}
	wsResponse2 = matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.WebsocketResponseType != matchserver.ResponseAsyncT ||
		wsResponse.ResponseAsync.MatchDetails.Color ==
			wsResponse2.ResponseAsync.MatchDetails.Color {
		t.Error("Expected the two players to have different colors")
	}
}

func TestWSReconnect(t *testing.T) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	u := "ws" + strings.TrimPrefix(serverMatchAndPlay.URL, "http")
	wsDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}
	wsDialer2 := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client2.Jar,
	}
	ws, _, err := wsDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	ws2, _, err := wsDialer2.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         matchserver.RequestAsync{Match: true},
	}
	ws.WriteJSON(&message)
	ws2.WriteJSON(&message)
	wsResponse := matchserver.WebsocketResponse{}
	wsResponse2 := matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.WebsocketResponseType != matchserver.ResponseAsyncT ||
		wsResponse.ResponseAsync.MatchDetails.Color ==
			wsResponse2.ResponseAsync.MatchDetails.Color {
		t.Error("Expected the two players to have different colors")
	}
	black := ws
	blackDialer := wsDialer
	white := ws2
	if wsResponse.ResponseAsync.MatchDetails.Color == model.White {
		black = ws2
		blackDialer = wsDialer2
		white = ws
	}
	err = black.Close()
	if err != nil {
		t.Error("Expected to close the black ws connection")
	}
	black, _, err = blackDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	playerResp, enemyResp := makeAsyncReq(
		matchserver.RequestAsync{Resign: true}, black, white)
	if !enemyResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	if !playerResp.ResponseAsync.GameOver {
		t.Error("Expected gameover")
	}
	white.Close()
	black.Close()
}

func makeAsyncReq(asyncReq matchserver.RequestAsync, player *websocket.Conn,
	enemy *websocket.Conn) (matchserver.WebsocketResponse,
	matchserver.WebsocketResponse) {
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestAsyncT,
		RequestAsync:         asyncReq,
	}
	player.WriteJSON(&message)
	playerResponse := matchserver.WebsocketResponse{}
	enemyResponse := matchserver.WebsocketResponse{}
	enemy.ReadJSON(&enemyResponse)
	if enemyResponse.ResponseAsync.GameOver {
		player.ReadJSON(&playerResponse)
	}
	return playerResponse, enemyResponse
}

func makeMove(x uint8, y uint8, moveX int8, moveY int8, player *websocket.Conn,
	enemy *websocket.Conn) (matchserver.WebsocketResponse,
	matchserver.WebsocketResponse) {
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestSyncT,
		RequestSync: model.MoveRequest{
			Position:  model.Position{File: x, Rank: y},
			Move:      model.Move{X: moveX, Y: moveY},
			PromoteTo: nil,
		},
	}
	player.WriteJSON(&message)
	playerResponse := matchserver.WebsocketResponse{}
	enemyResponse := matchserver.WebsocketResponse{}
	player.ReadJSON(&playerResponse)
	if playerResponse.ResponseSync.MoveSuccess {
		enemy.ReadJSON(&enemyResponse)
	}
	return playerResponse, enemyResponse
}

func startSession(client *http.Client, username string) {
	credentialsBuf := new(bytes.Buffer)
	credentials := gateway.Credentials{Username: username}
	json.NewEncoder(credentialsBuf).Encode(credentials)
	resp, err := client.Post(serverSession.URL, "application/json", credentialsBuf)
	serverSessionURL, _ := url.Parse(serverSession.URL)
	serverMatchAndPlayURL, _ := url.Parse(serverMatchAndPlay.URL)
	// Ensure that the various test handler URLs get passed the session cookie
	// by the client.
	client.Jar.SetCookies(serverMatchAndPlayURL, client.Jar.Cookies(serverSessionURL))
	if err == nil {
		defer resp.Body.Close()
	}
}
