all: buildrpc buildweb
	go install ./...
	go test ./...

webclient_package := github.com/Ekotlikoff/gochess/internal/client/web
run_local_package := github.com/Ekotlikoff/gochess/cmd/webserver

buildweb:
	GOARCH=wasm GOOS=js go build \
		   -o $(GOPATH)/src/gochess/internal/server/frontend/static/lib.wasm \
		   -tags webclient $(webclient_package)

buildrpc:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/chessengine.proto

runweb: buildweb
	go run $(run_local_package)
