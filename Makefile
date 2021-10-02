webclient_package := github.com/Ekotlikoff/gochess/internal/client/web
run_local_package := github.com/Ekotlikoff/gochess/cmd/gochess

all: vet proto web test testrace

clean:
	go clean -i github.com/Ekotlikoff/gochess/...

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/chessengine.proto

vet:
	./vet.sh -install
	./vet.sh

test:
	go test -cpu 1,4 -timeout 7m github.com/Ekotlikoff/gochess/...

testrace:
	go test -race -cpu 1,4 -timeout 7m github.com/Ekotlikoff/gochess/...

web:
	GOARCH=wasm GOOS=js go build \
		-o ~/bin/gochessclient.wasm \
		-tags webclient $(webclient_package)

runweb: web
	go run $(run_local_package)

.PHONY: \
	all \
	clean \
	proto \
	runweb \
	test \
	testrace \
	vet \
