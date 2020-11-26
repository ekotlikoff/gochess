package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Ekotlikoff/gochess/internal/model"
	"github.com/Ekotlikoff/gochess/internal/server/match"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var (
	debug              bool   = false
	ctp                string = "application/json"
	serverMatch        *httptest.Server
	serverSession      *httptest.Server
	serverSync         *httptest.Server
	serverAsync        *httptest.Server
	serverMatchTimeout *httptest.Server
)

func init() {
	serverSession = httptest.NewServer(http.HandlerFunc(StartSession))
	serverAsync = httptest.NewServer(http.HandlerFunc(AsyncHandler))
	serverSync = httptest.NewServer(http.HandlerFunc(SyncHandler))
	matchingServer := matchserver.NewMatchingServer()
	serverMatch = httptest.NewServer(
		createSearchForMatchHandler(&matchingServer))
	exitChan := make(chan bool, 1)
	close(exitChan)
	matchingServer.StartMatchServers(10, exitChan)
	timeoutMatchingServer := matchserver.NewMatchingServer()
	serverMatchTimeout = httptest.NewServer(
		createSearchForMatchHandler(&timeoutMatchingServer))
	generator := func(
		black *matchserver.Player, white *matchserver.Player,
	) matchserver.Match {
		return matchserver.NewMatch(black, white, 100)
	}
	timeoutMatchingServer.StartCustomMatchServers(10, generator, exitChan)
	SetQuiet()
}

func TestHTTPServerStartSession(t *testing.T) {
	if debug {
		fmt.Println("Test StartSession")
	}
	credentialsBuf := new(bytes.Buffer)
	credentials := Credentials{"my_username"}
	json.NewEncoder(credentialsBuf).Encode(credentials)
	resp, err := http.Post(serverSession.URL, ctp, credentialsBuf)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Println(body)
		fmt.Println(resp.Cookies())
		fmt.Println(err)
	}
	if resp.StatusCode != 200 || err != nil || len(resp.Cookies()) == 0 {
		t.Error("Expected cookies")
	}
}

func TestHTTPServerStartSessionError(t *testing.T) {
	if debug {
		fmt.Println("Test StartSessionError")
	}
	resp, _ := http.Post(serverSession.URL, "application/json", bytes.NewBuffer([]byte("error")))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Println(string(body))
		fmt.Println(resp.Cookies())
	}
	if resp.StatusCode == 200 || len(resp.Cookies()) == 1 {
		t.Error("Expected failure")
	}
}

func TestHTTPServerStartSessionNoUsername(t *testing.T) {
	if debug {
		fmt.Println("Test StartSessionNoUsername")
	}
	requestBody, _ := json.Marshal(map[string]string{
		"not_username": "my_username",
	})
	resp, _ := http.Post(serverSession.URL, "application/json", bytes.NewBuffer(requestBody))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Println(string(body))
		fmt.Println(resp.Cookies())
	}
	if resp.StatusCode == 200 || len(resp.Cookies()) == 1 ||
		!strings.HasPrefix(string(body), "Missing username") {
		t.Error("Expected failure")
	}
}

func TestHTTPServerMatch(t *testing.T) {
	if debug {
		fmt.Println("Test Match")
	}
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	wait := make(chan struct{})
	var resp *http.Response
	go func() { resp, _ = client.Get(serverMatch.URL); close(wait) }()
	resp2, err2 := client2.Get(serverMatch.URL)
	<-wait
	defer resp.Body.Close()
	defer resp2.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	body2, _ := ioutil.ReadAll(resp2.Body)
	if debug {
		fmt.Println(string(body))
		fmt.Println(resp.StatusCode)
		fmt.Println(string(body2))
		fmt.Println(resp2.StatusCode)
	}
	if err2 != nil || (!strings.HasPrefix(string(body), "{\"Color\":1,\"Opp") &&
		!strings.HasPrefix(string(body2), "{\"Color\":1,\"Opp")) {
		t.Error("Expected match got ", string(body))
	}
}

func TestHTTPServerCheckmate(t *testing.T) {
	if debug {
		fmt.Println("Test Checkmate")
	}
	black, white, blackName, _ := createMatch(serverMatch)
	sendMove(white, serverSync, 2, 1, 0, 2)
	resp, _ := black.Get(serverSync.URL)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body),
		"{\"Position\":{\"File\":2,\"Rank\":1},\"Move\":{\"X\":0,\"Y\":2}}") {
		t.Error("Expected opponent's move got ", string(body))
	}
	sendMove(black, serverSync, 4, 6, 0, -2)
	resp, _ = white.Get(serverSync.URL)
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body),
		"{\"Position\":{\"File\":4,\"Rank\":6},\"Move\":{\"X\":0,\"Y\":-2}}") {
		t.Error("Expected opponent's move got ", string(body))
	}
	sendMove(white, serverSync, 2, 3, 0, 1)
	sendMove(black, serverSync, 3, 7, 4, -4)
	sendMove(white, serverSync, 2, 4, 0, 1)
	sendMove(black, serverSync, 5, 7, -3, -3)
	sendMove(white, serverSync, 2, 5, -1, 1)
	sendMove(black, serverSync, 7, 3, -2, -2)
	resp, _ = white.Get(serverAsync.URL)
	responseAsync := matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.GameOver || responseAsync.Winner != blackName {
		t.Error("Expected gameover got ", responseAsync)
	}
	if debug {
		fmt.Println("Success Checkmate")
	}
}

func TestHTTPServerDraw(t *testing.T) {
	if debug {
		fmt.Println("Test Draw")
	}
	black, white, _, _ := createMatch(serverMatch)
	sendMove(white, serverSync, 2, 1, 0, 2)
	payloadBuf := new(bytes.Buffer)
	requestAsync := matchserver.RequestAsync{RequestToDraw: true}
	json.NewEncoder(payloadBuf).Encode(requestAsync)
	white.Post(serverAsync.URL, ctp, payloadBuf)
	resp, _ := black.Get(serverAsync.URL)
	responseAsync := matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.RequestToDraw || responseAsync.GameOver {
		t.Error("Expected draw request from white got ", responseAsync)
	}
	json.NewEncoder(payloadBuf).Encode(requestAsync)
	black.Post(serverAsync.URL, ctp, payloadBuf)
	resp, _ = white.Get(serverAsync.URL)
	responseAsync = matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.GameOver || responseAsync.Winner != "" ||
		!responseAsync.Draw || responseAsync.Resignation {
		t.Error("Expected gameover got ", responseAsync)
	}
}

func TestHTTPServerResign(t *testing.T) {
	if debug {
		fmt.Println("Test Resign")
	}
	black, white, blackName, _ := createMatch(serverMatch)
	sendMove(white, serverSync, 2, 1, 0, 2)
	payloadBuf := new(bytes.Buffer)
	requestAsync := matchserver.RequestAsync{Resign: true}
	json.NewEncoder(payloadBuf).Encode(requestAsync)
	white.Post(serverAsync.URL, ctp, payloadBuf)
	resp, _ := black.Get(serverAsync.URL)
	responseAsync := matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.GameOver || responseAsync.Winner != blackName ||
		!responseAsync.Resignation {
		t.Error("Expected gameover got ", responseAsync)
	}
}

func TestHTTPServerTimeout(t *testing.T) {
	if debug {
		fmt.Println("Test Timeout")
	}
	black, white, blackName, _ := createMatch(serverMatchTimeout)
	sendMove(white, serverSync, 2, 1, 0, 2)
	sendMove(black, serverSync, 2, 6, 0, -2)
	resp, _ := black.Get(serverAsync.URL)
	responseAsync := matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.GameOver || responseAsync.Winner != blackName ||
		!responseAsync.Timeout {
		t.Error("Expected timeout got ", responseAsync)
	}
}

func createMatch(testMatchServer *httptest.Server) (
	black *http.Client, white *http.Client, blackName string, whiteName string,
) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	startSession(client, "player1")
	startSession(client2, "player2")
	wait := make(chan struct{})
	var resp *http.Response
	go func() { resp, _ = client.Get(testMatchServer.URL); close(wait) }()
	resp2, _ := client2.Get(testMatchServer.URL)
	<-wait
	defer resp.Body.Close()
	defer resp2.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	black = client
	blackName = "player1"
	whiteName = "player2"
	white = client2
	if strings.Contains(string(body), "1") {
		black = client2
		blackName = "player2"
		whiteName = "player1"
		white = client
	}
	return black, white, blackName, whiteName
}

func sendMove(client *http.Client, serverSync *httptest.Server, x, y,
	moveX, moveY int) {
	movePayloadBuf := new(bytes.Buffer)
	moveRequest := model.MoveRequest{
		model.Position{uint8(x), uint8(y)}, model.Move{int8(moveX), int8(moveY)},
	}
	json.NewEncoder(movePayloadBuf).Encode(moveRequest)
	resp, err := client.Post(serverSync.URL, "application/json", movePayloadBuf)
	if err == nil {
		defer resp.Body.Close()
	}
}

func startSession(client *http.Client, username string) {
	credentialsBuf := new(bytes.Buffer)
	credentials := Credentials{username}
	json.NewEncoder(credentialsBuf).Encode(credentials)
	resp, err := client.Post(serverSession.URL, "application/json", credentialsBuf)
	serverSessionURL, _ := url.Parse(serverSession.URL)
	serverMatchURL, _ := url.Parse(serverMatch.URL)
	serverSyncURL, _ := url.Parse(serverSync.URL)
	serverAsyncURL, _ := url.Parse(serverAsync.URL)
	// Ensure that the various test handler URLs get passed the session cookie
	// by the client.
	client.Jar.SetCookies(serverMatchURL, client.Jar.Cookies(serverSessionURL))
	client.Jar.SetCookies(serverSyncURL, client.Jar.Cookies(serverSessionURL))
	client.Jar.SetCookies(serverAsyncURL, client.Jar.Cookies(serverSessionURL))
	if err == nil {
		defer resp.Body.Close()
	}
}
