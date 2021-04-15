package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	debug         bool   = false
	ctp           string = "application/json"
	serverSession *httptest.Server
)

func init() {
	serverSession = httptest.NewServer(http.HandlerFunc(StartSession))
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
	resp, err := http.Post(serverSession.URL, "application/json", bytes.NewBuffer([]byte("error")))
	if err != nil {
		t.Error("Did not expect an error")
	}
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
	resp, err := http.Post(serverSession.URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Error("Did not expect an error")
	}
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
