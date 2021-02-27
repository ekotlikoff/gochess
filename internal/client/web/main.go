// +build webclient

package main

import (
	_ "embed"
	"encoding/json"
	"github.com/Ekotlikoff/gochess/internal/model"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"syscall/js"
	"time"
)

var (
	//go:embed config.json
	config []byte
	quiet  bool = false
)

type Configuration struct {
	BackendType   BackendType
	ClientTimeout string
}

func main() {
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	config := loadConfig()
	done := make(chan struct{}, 0)
	game := model.NewGame()
	jar, _ := cookiejar.New(&cookiejar.Options{})
	clientTimeout, _ := time.ParseDuration(config.ClientTimeout)
	client := &http.Client{Jar: jar, Timeout: clientTimeout}
	wsDialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
		Jar:              client.Jar,
	}
	clientModel := ClientModel{
		game: &game, playerColor: model.White,
		document: js.Global().Get("document"),
		board: js.Global().Get("document").Call(
			"getElementById", "board-layout-chessboard"),
		backendType: config.BackendType, client: client, wsDialer: wsDialer,
		origin: js.Global().Get("window").Get("location").Get("host").String(),
	}
	clientModel.initController()
	clientModel.initStyle()
	clientModel.viewInitBoard(clientModel.playerColor)
	<-done
}

func loadConfig() Configuration {
	configuration := Configuration{}
	err := json.Unmarshal(config, &configuration)
	if err != nil {
		log.Println("ERROR:", err)
	}
	return configuration
}
