package httpserver

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

func init() {
	matchingServer := matchserver.NewMatchingServer()
	exitChan := make(chan bool, 1)
	exitChan <- true
	matchingServer.Serve(1, exitChan)
	go Serve(&matchingServer, nil, false)
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
	resp, _ := http.Post(
		"http://localhost:80/", "text/plain", bytes.NewBuffer([]byte("error")),
	)
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
	requestBody, _ := json.Marshal(map[string]string{
		"not_username": "my_username",
	})
	resp, _ := http.Post(
		"http://localhost:80/", "text/plain", bytes.NewBuffer(requestBody),
	)
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
	requestBody, _ := json.Marshal(map[string]string{
		"username": "player1",
	})
	requestBody2, _ := json.Marshal(map[string]string{
		"username": "player2",
	})
	jar, _ := cookiejar.New(&cookiejar.Options{})
	jar2, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{
		Jar: jar,
	}
	client2 := &http.Client{
		Jar: jar2,
	}
	resp, err := client.Post(
		"http://localhost:80/", "text/plain", bytes.NewBuffer(requestBody),
	)
	resp2, err2 := client2.Post(
		"http://localhost:80/", "text/plain", bytes.NewBuffer(requestBody2),
	)
	defer resp.Body.Close()
	defer resp2.Body.Close()
	wait := make(chan struct{})
	go func() { resp, err = client.Get("http://localhost:80/match"); close(wait) }()
	resp2, err2 = client2.Get("http://localhost:80/match")
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

//for len(matchingServer.LiveMatches()) == 0 {
//}
//liveMatch := matchingServer.LiveMatches()[0]
