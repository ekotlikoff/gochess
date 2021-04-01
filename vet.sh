#!/usr/bin/env bash

set -ex  # Exit on error; debugging enabled.
set -o pipefail  # Fail a pipe if any sub-command fails.

# not makes sure the command passed to it does not exit with a return code of 0.
not() {
    # This is required instead of the earlier (! $COMMAND) because subshells and
    # pipefail don't work the same on Darwin as in Linux.
    ! "$@"
}

fail_on_output() {
    tee /dev/stderr | not read
}

PATH="${HOME}/go/bin:${GOROOT}/bin:${PATH}"
go version

if [[ "${TRAVIS}" = "true" ]]; then
    PROTOBUF_VERSION=3.14.0
    PROTOC_FILENAME=protoc-${PROTOBUF_VERSION}-linux-x86_64.zip
    pushd /home/travis
    wget https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/${PROTOC_FILENAME}
    unzip ${PROTOC_FILENAME}
    bin/protoc --version
    popd
fi

# - gofmt, goimports, golint (with exceptions for generated code), go vet.
#gofmt -s -d -l . 2>&1 | fail_on_output
#goimports -l . 2>&1 | not grep -vE "\.pb\.go"
#golint ./... 2>&1 | not grep -vE "/testv3\.pb\.go:"
#go vet -all ./...

PATH="/home/travis/bin:${PATH}" make proto && \
    git status --porcelain 2>&1 | fail_on_output || \
    (git status; git --no-pager diff; exit 1)
