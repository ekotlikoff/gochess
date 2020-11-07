package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gochess/internal/server/match"
	"io/ioutil"
	"net/http"
	"testing"
)

var debug bool = true

func init() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(1, exitChan)
	go Serve(&matchingServer)
}

func TestHTTPServerStartSession(t *testing.T) {
	requestBody, err := json.Marshal(map[string]string{
		"username": "my_username",
	})
	resp, err := http.Post(
		"http://localhost:80/", "text/plain", bytes.NewBuffer(requestBody),
	)
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
	requestBody, err := json.Marshal(map[string]string{
		"not_username": "my_username",
	})
	resp, rerr := http.Post(
		"http://localhost:80/", "text/plain", bytes.NewBuffer(requestBody),
	)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Println(body)
		fmt.Println(resp.Cookies())
		fmt.Println(err)
		fmt.Println(rerr)
		fmt.Println(resp)
	}
	if resp.StatusCode == 200 || len(resp.Cookies()) == 1 {
		t.Error("Expected no cookies")
	}
}

//for len(matchingServer.LiveMatches()) == 0 {
//}
//liveMatch := matchingServer.LiveMatches()[0]
