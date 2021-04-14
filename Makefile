all: proto web
	go install ./...
	go test ./...

webclient_package := github.com/Ekotlikoff/gochess/internal/client/web
run_local_package := github.com/Ekotlikoff/gochess/cmd/webserver

web:
	GOARCH=wasm GOOS=js go build \
		   -o $(GOPATH)/src/gochess/internal/server/frontend/static/lib.wasm \
		   -tags webclient $(webclient_package)

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/chessengine.proto

testrace:
	go test -race -cpu 1,4 -timeout 7m ./...

runweb: web
	go run $(run_local_package)
