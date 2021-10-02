A chess server where players can match and play with others using their web browser.

## Build
* Install protoc https://grpc.io/docs/protoc-installation/
* Install Golang
* Run `make`

## Observability
* Run jaeger all-in-one to collect traces
* Run prometheus and scrape /metrics to collect metrics

## Components
* Web client compiled from Golang with WebAssembly
* Frontend gateway which serves static files and proxies calls to a backend server
* Interchangeable HTTP and WebSocket web servers communicate with the web client and forwards requests to the match server
* Client agnostic match server orchestrates a match
* Backend chess matching server with the ability to match a player with a remote chess engine
* Chess model for pieces, moves, the board, and a game

## Usage as a package
* `go get github.com/Ekotlikoff/gochess/pkg/chessserver`
*	`GOARCH=wasm GOOS=js go build -o ~/bin/gochessclient.wasm -tags webclient github.com/Ekotlikoff/gochess/internal/client/web`
* `go chessserver.RunServer()`
