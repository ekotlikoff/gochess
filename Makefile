all:
	go test ./...

buildweb:
	GOARCH=wasm GOOS=js go build -o \
		   $(GOPATH)/src/gochess/cmd/webserver/assets/lib.wasm \
		   -tags webclient gochess/internal/server/http/webclient

runweb: buildweb
	go run gochess/cmd/webserver
