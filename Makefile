all:
	go install ./...
	go test ./...

server := http://192.168.1.166:8000/
webclient_package := github.com/Ekotlikoff/gochess/internal/server/http/webclient

buildweb:
	GOARCH=wasm GOOS=js go build \
           -ldflags "-X $(webclient_package).server=$(server)" \
		   -o $(GOPATH)/src/gochess/cmd/webserver/assets/lib.wasm \
		   -tags webclient $(webclient_package)

runweb: buildweb
	go run github.com/Ekotlikoff/gochess/cmd/webserver
