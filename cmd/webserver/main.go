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
	// Configuration is a struct that configures the chess server
	Configuration struct {
		BackendType             BackendType
		EnableBotMatching       bool
		EngineConnectionTimeout string
		EngineAddr              string
		MaxMatchingDuration     string
	}
	// BackendType represents different types of backends
	BackendType string
)

const (
	// HTTPBackend type
	HTTPBackend = BackendType("http")
	// WebsocketBackend type
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
	if config.BackendType == HTTPBackend {
		go httpserver.Serve(&matchingServer, gateway.SessionCache, 8001, nil,
			false)
	} else if config.BackendType == WebsocketBackend {
		go websocketserver.Serve(&matchingServer, gateway.SessionCache, 8002,
			nil, false)
	}
	httpserverURL, _ := url.Parse("http://localhost:8001")
	websocketURL, _ := url.Parse("http://localhost:8002")
	gateway.Serve(httpserverURL, websocketURL, 8000, nil, false)
}

func loadConfig() Configuration {
	configuration := Configuration{}
	err := json.Unmarshal(config, &configuration)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	return configuration
}
