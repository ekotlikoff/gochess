package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochess/internal/server/match"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
)

var debug bool = false
var uri string = "http://localhost:8000/"

func init() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(10, exitChan)
	go Serve(&matchingServer, 8000, nil, true)
}

func TestHTTPServerStartSession(t *testing.T) {
	if debug {
		fmt.Println("Test StartSession")
	}
	requestBody, err := json.Marshal(map[string]string{
		"username": "my_username",
	})
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(requestBody))
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
	resp, _ := http.Post(uri, "application/json", bytes.NewBuffer([]byte("error")))
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
	resp, _ := http.Post(uri, "application/json", bytes.NewBuffer(requestBody))
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
	requestBody, _ := json.Marshal(map[string]string{"username": "player1"})
	requestBody2, _ := json.Marshal(map[string]string{"username": "player2"})
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	resp, err := client.Post(
		uri, "application/json", bytes.NewBuffer(requestBody))
	resp2, err2 := client2.Post(
		uri, "application/json", bytes.NewBuffer(requestBody2))
	defer resp.Body.Close()
	defer resp2.Body.Close()
	wait := make(chan struct{})
	go func() { resp, err = client.Get(uri + "match"); close(wait) }()
	resp2, err2 = client2.Get(uri + "match")
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
	if err != nil || err2 != nil || !strings.HasPrefix(string(body), "Matched!") ||
		!strings.HasPrefix(string(body2), "Matched!") {
		t.Error("Expected match got ", string(body))
	}
}

func TestHTTPServerCheckmate(t *testing.T) {
	if debug {
		fmt.Println("Test Checkmate")
	}
	black, white, blackName, _ := createMatch(uri)
	ctp := "application/json"
	reqMoveBody, _ := json.Marshal(map[string]string{"move": "(2,1)(0,2)"})
	white.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	resp, _ := black.Get(uri + "sync")
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "{2 1} {0 2}") {
		t.Error("Expected opponent's move got ", string(body))
	}
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(4,6)(0,-2)"})
	black.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	resp, _ = white.Get(uri + "sync")
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "{4 6} {0 -2}") {
		t.Error("Expected opponent's move got ", string(body))
	}
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(2,3)(0,1)"})
	white.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(3,7)(4,-4)"})
	black.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(2,4)(0,1)"})
	white.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(5,7)(-3,-3)"})
	black.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(2,5)(-1,1)"})
	white.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(7,3)(-2,-2)"})
	black.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	resp, _ = white.Get(uri + "async")
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
	black, white, _, _ := createMatch(uri)
	ctp := "application/json"
	reqMoveBody, _ := json.Marshal(map[string]string{"move": "(2,1)(0,2)"})
	white.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	payloadBuf := new(bytes.Buffer)
	requestAsync := matchserver.RequestAsync{RequestToDraw: true}
	json.NewEncoder(payloadBuf).Encode(requestAsync)
	white.Post(uri+"async", ctp, payloadBuf)
	resp, _ := black.Get(uri + "async")
	responseAsync := matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.RequestToDraw || responseAsync.GameOver {
		t.Error("Expected draw request from white got ", responseAsync)
	}
	json.NewEncoder(payloadBuf).Encode(requestAsync)
	black.Post(uri+"async", ctp, payloadBuf)
	resp, _ = white.Get(uri + "async")
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
	black, white, blackName, _ := createMatch(uri)
	ctp := "application/json"
	reqMoveBody, _ := json.Marshal(map[string]string{"move": "(2,1)(0,2)"})
	white.Post(uri+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	payloadBuf := new(bytes.Buffer)
	requestAsync := matchserver.RequestAsync{Resign: true}
	json.NewEncoder(payloadBuf).Encode(requestAsync)
	white.Post(uri+"async", ctp, payloadBuf)
	resp, _ := black.Get(uri + "async")
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
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	generator := func(
		black *matchserver.Player, white *matchserver.Player,
	) matchserver.Match {
		return matchserver.NewMatch(black, white, 100)
	}
	matchingServer.ServeCustomMatch(10, generator, exitChan)
	go Serve(&matchingServer, 8001, nil, true)
	uri_timeout := "http://localhost:8001/"
	black, white, blackName, _ := createMatch(uri_timeout)
	ctp := "application/json"
	reqMoveBody, _ := json.Marshal(map[string]string{"move": "(2,1)(0,2)"})
	white.Post(uri_timeout+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	reqMoveBody, _ = json.Marshal(map[string]string{"move": "(2,6)(0,-2)"})
	black.Post(uri_timeout+"sync", ctp, bytes.NewBuffer(reqMoveBody))
	resp, _ := black.Get(uri_timeout + "async")
	responseAsync := matchserver.ResponseAsync{}
	json.NewDecoder(resp.Body).Decode(&responseAsync)
	if !responseAsync.GameOver || responseAsync.Winner != blackName ||
		!responseAsync.Timeout {
		t.Error("Expected timeout got ", responseAsync)
	}
}

func createMatch(uri string) (
	black *http.Client, white *http.Client, blackName string, whiteName string,
) {
	requestBody, _ := json.Marshal(map[string]string{"username": "player1"})
	requestBody2, _ := json.Marshal(map[string]string{"username": "player2"})
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{Jar: jar}
	client2 := &http.Client{Jar: jar2}
	resp, _ := client.Post(uri, "application/json", bytes.NewBuffer(requestBody))
	resp2, _ := client2.Post(
		uri, "application/json", bytes.NewBuffer(requestBody2))
	defer resp.Body.Close()
	defer resp2.Body.Close()
	wait := make(chan struct{})
	go func() { resp, _ = client.Get(uri + "match"); close(wait) }()
	resp2, _ = client2.Get(uri + "match")
	<-wait
	defer resp.Body.Close()
	defer resp2.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	black = client
	blackName = "player1"
	whiteName = "player2"
	white = client2
	if strings.Contains(string(body), "color=1") {
		black = client2
		blackName = "player2"
		whiteName = "player1"
		white = client
	}
	return black, white, blackName, whiteName
}
