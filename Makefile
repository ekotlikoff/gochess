all: buildweb
	go install ./...
	go test ./...

webclient_package := github.com/Ekotlikoff/gochess/internal/client/web
run_local_package := github.com/Ekotlikoff/gochess/cmd/webserver

buildweb:
	GOARCH=wasm GOOS=js go build \
		   -o $(GOPATH)/src/gochess/internal/server/frontend/static/lib.wasm \
		   -tags webclient $(webclient_package)

runweb: buildweb
	go run $(run_local_package)
