A chess server where players can match and play with others using their web browser.

## Build
* Install [protoc](https://grpc.io/docs/protoc-installation/)
* Install [Golang](https://golang.org/doc/install)
* Run `make`

## Observability
* Run jaeger all-in-one to collect traces
* Run prometheus and scrape /metrics to collect metrics

## Components
* [Web client](internal/client/web/main.go) compiled from Golang with WebAssembly
* [Frontend gateway](internal/server/frontend/gateway.go) which serves static files and proxies calls to a backend server
* Interchangeable [HTTP](internal/server/backend/http/webserver.go) and [WebSocket](internal/server/backend/websocket/websocketserver.go) backend servers communicate with the web client and relay requests/responses to/from the match server via player channels
* Client agnostic [match server](internal/server/backend/match/matchserver.go) orchestrates a [match](internal/server/backend/match/match.go)
* [Matching server](internal/server/backend/match/match.go) with the ability to match a player with a [remote chess engine](https://github.com/ekotlikoff/rustchess) via an [engine client](internal/server/backend/match/engine_client.go)
* [Chess model](internal/model/model.go) for [pieces](internal/model/piece.go), [moves](internal/model/move.go), the [board](internal/model/model.go), and a [game](internal/model/game.go)

## Usage as a package
* `go get github.com/ekotlikoff/gochess/pkg/chessserver`
*	`GOARCH=wasm GOOS=js go build -o ~/bin/gochessclient.wasm -tags webclient github.com/ekotlikoff/gochess/internal/client/web`
* `go chessserver.RunServer()`
