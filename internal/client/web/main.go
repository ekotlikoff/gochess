//go:build webclient

package main

import (
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"syscall/js"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"
)

var (
	//go:embed config.json
	config []byte
	quiet  bool = false
)

type Configuration struct {
	BackendType   BackendType
	TLS           bool
	ClientTimeout string
}

func main() {
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	config := loadConfig()
	done := make(chan struct{})
	game := model.NewGame()
	jar, _ := cookiejar.New(&cookiejar.Options{})
	clientTimeout, _ := time.ParseDuration(config.ClientTimeout)
	client := &http.Client{Jar: jar, Timeout: clientTimeout}
	clientModel := ClientModel{
		game: game, playerColor: model.White,
		document: js.Global().Get("document"),
		board: js.Global().Get("document").Call(
			"getElementById", "board-layout-chessboard"), tls: config.TLS,
		backendType: config.BackendType, client: client, gameType: Local,
		origin: js.Global().Get("window").Get("location").Get("host").String(),
	}
	clientModel.initController()
	clientModel.initStyle()
	clientModel.viewInitBoard(clientModel.playerColor)
	clientModel.checkForSession()
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
