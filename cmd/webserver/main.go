package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	httpserver "github.com/Ekotlikoff/gochess/internal/server/backend/http"
	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	websocketserver "github.com/Ekotlikoff/gochess/internal/server/backend/websocket"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
)

//go:embed config.json
var config []byte

type (
	Configuration struct {
		BackendType             BackendType
		EnableBotMatching       bool
		EngineConnectionTimeout string
		EngineAddr              string
		MaxMatchingDuration     string
	}
	BackendType string
)

const (
	HttpBackend      = BackendType("http")
	WebsocketBackend = BackendType("websocket")
)

func main() {
	config := loadConfig()
	engineConnTimeout, _ := time.ParseDuration(config.EngineConnectionTimeout)
	maxMatchingDuration, _ := time.ParseDuration(config.MaxMatchingDuration)
	var matchingServer matchserver.MatchingServer
	if !config.EnableBotMatching {
		matchingServer = matchserver.NewMatchingServer()
	} else {
		matchingServer = matchserver.NewMatchingServerWithEngine(
			config.EngineAddr, maxMatchingDuration, engineConnTimeout)
	}
	exitChan := make(chan bool, 1)
	go matchingServer.StartMatchServers(10, exitChan)
	if config.BackendType == HttpBackend {
		go httpserver.Serve(&matchingServer, gateway.SessionCache, 8001, nil,
			false)
	} else if config.BackendType == WebsocketBackend {
		go websocketserver.Serve(&matchingServer, gateway.SessionCache, 8002,
			nil, false)
	}
	httpserverUrl, _ := url.Parse("http://localhost:8001")
	websocketUrl, _ := url.Parse("http://localhost:8002")
	gateway.Serve(httpserverUrl, websocketUrl, 8000, nil, false)
}

func loadConfig() Configuration {
	configuration := Configuration{}
	err := json.Unmarshal(config, &configuration)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	return configuration
}
