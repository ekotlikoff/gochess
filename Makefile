all: buildweb
	go install ./...
	go test ./...

buildweb:
	GOARCH=wasm GOOS=js go build -o \
		   $(GOPATH)/src/gochess/cmd/webserver/assets/lib.wasm \
		   -tags webclient github.com/Ekotlikoff/gochess/internal/server/http/webclient

runweb: buildweb
	go run github.com/Ekotlikoff/gochess/cmd/webserver
