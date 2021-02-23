all: buildweb
	go install ./...
	go test ./...

server := http://192.168.1.166:8000/
webclient_package := github.com/Ekotlikoff/gochess/internal/client/web
webserver_cmd_package := github.com/Ekotlikoff/gochess/cmd/webserver

buildweb:
	GOARCH=wasm GOOS=js go build \
		   -o $(GOPATH)/src/gochess/internal/server/api/http/static/lib.wasm \
		   -tags webclient $(webclient_package)

runweb: buildweb
	go run $(webserver_cmd_package)
