package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	httpserver "github.com/Ekotlikoff/gochess/internal/server/backend/http"
	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	websocketserver "github.com/Ekotlikoff/gochess/internal/server/backend/websocket"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

//go:embed config.json
var config []byte

type (
	// Configuration is a struct that configures the chess server
	Configuration struct {
		ServiceName             string
		Environment             string
		BackendType             BackendType
		EnableBotMatching       bool
		EngineConnectionTimeout string
		EngineAddr              string
		GatewayPort             int
		HTTPPort                int
		WSPort                  int
		MaxMatchingDuration     string
		LogFile                 string
		EnableTracing           bool
		Quiet                   bool
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
	configureLogging(config)
	if config.EnableTracing {
		closer := configureTracing(config)
		if closer != nil {
			defer closer.Close()
		}
	}
	startChessServer(config)
}

func startChessServer(config Configuration) {
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
		go httpserver.Serve(&matchingServer, config.HTTPPort)
	} else if config.BackendType == WebsocketBackend {
		go websocketserver.Serve(&matchingServer, config.WSPort)
	}
	httpserverURL, _ := url.Parse("http://localhost:" +
		strconv.Itoa(config.HTTPPort))
	websocketURL, _ := url.Parse("http://localhost:" +
		strconv.Itoa(config.WSPort))
	gateway.Serve(httpserverURL, websocketURL, config.GatewayPort)
}

func configureLogging(config Configuration) {
	if config.LogFile != "" {
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(file)
	}
	if config.Quiet {
		log.SetOutput(ioutil.Discard)
	}
}

func configureTracing(config Configuration) io.Closer {
	// Recommended configuration for production.
	cfg := jaegercfg.Configuration{}
	if config.Environment == "local" {
		cfg = jaegercfg.Configuration{
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
		}
	}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	closer, err := cfg.InitGlobalTracer(
		config.ServiceName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return closer
	}
	return closer
}

func loadConfig() Configuration {
	configuration := Configuration{}
	err := json.Unmarshal(config, &configuration)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	return configuration
}
