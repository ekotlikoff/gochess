package websocketserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Ekotlikoff/gochess/internal/model"
	"github.com/Ekotlikoff/gochess/internal/server/match"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var (
	debug              bool = false
	serverSession      *httptest.Server
	serverMatchAndPlay *httptest.Server
)

func init() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	close(exitChan)
	matchingServer.StartMatchServers(10, exitChan)
	serverSession = httptest.NewServer(http.HandlerFunc(StartSession))
	serverMatchAndPlay = httptest.NewServer(http.Handler(createMatchAndPlayHandler(&matchingServer)))
}

func TestWSMatch(t *testing.T) {
	if debug {
		fmt.Println("Test StartSession")
	}
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
	wsResponse := matchserver.WebsocketResponse{}
	wsResponse2 := matchserver.WebsocketResponse{}
	ws.ReadJSON(&wsResponse)
	ws2.ReadJSON(&wsResponse2)
	if wsResponse.MatchedResponse.Color == wsResponse2.MatchedResponse.Color {
		t.Error("Expected the two players to have different colors")
	}
	black := ws
	white := ws2
	if wsResponse.MatchedResponse.Color == model.White {
		black = ws2
		white = ws
	}
	message := matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestSyncT,
		RequestSync:          model.MoveRequest{},
	}
	white.WriteJSON(&message)
	white.ReadJSON(&wsResponse)
	if wsResponse.ResponseSync.MoveSuccess {
		t.Error("Expected invalid move response")
	}
	message = matchserver.WebsocketRequest{
		WebsocketRequestType: matchserver.RequestSyncT,
		RequestSync: model.MoveRequest{
			model.Position{0, 1}, model.Move{0, 2}},
	}
	white.WriteJSON(&message)
	white.ReadJSON(&wsResponse)
	if !wsResponse.ResponseSync.MoveSuccess {
		t.Error("Expected valid move response")
	}
	black.ReadJSON(&wsResponse)
	if wsResponse.OpponentPlayedMove.Move.Y != 2 {
		t.Error("Expected opponent's move")
	}
}

func startSession(client *http.Client, username string) {
	credentialsBuf := new(bytes.Buffer)
	credentials := Credentials{username}
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
