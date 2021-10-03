//go:build tools

// This package exists to cause `go mod` and `go get` to believe these tools
// are dependencies, even though they are not runtime dependencies.

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
