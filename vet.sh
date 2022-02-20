#!/usr/bin/env bash
# Credit to https://github.com/grpc/grpc-go/blob/master/vet.sh

set -ex  # Exit on error; debugging enabled.
set -o pipefail  # Fail a pipe if any sub-command fails.

# not makes sure the command passed to it does not exit with a return code of 0.
not() {
    # This is required instead of the earlier (! $COMMAND) because subshells and
    # pipefail don't work the same on Darwin as in Linux.
    ! "$@"
}

die() {
  echo "$@" >&2
  exit 1
}

fail_on_output() {
    tee /dev/stderr | not read
}

# Check to make sure it's safe to modify the user's git repo.
git status --porcelain | fail_on_output

# Undo any edits made by this script.
cleanup() {
  git reset --hard HEAD
}
trap cleanup EXIT

PATH="${HOME}/go/bin:${GOROOT}/bin:${PATH}"
go version

if [[ "$1" = "-install" ]]; then
     # Install the pinned versions as defined in module tools.
    pushd ./test/tools
    go install \
      golang.org/x/lint/golint \
      golang.org/x/tools/cmd/goimports \
      honnef.co/go/tools/cmd/staticcheck \
      github.com/client9/misspell/cmd/misspell \
      google.golang.org/protobuf/cmd/protoc-gen-go \
      google.golang.org/grpc/cmd/protoc-gen-go-grpc
    popd

    if [[ "${TRAVIS}" = "true" || "${CODEBUILD}" = 1 ]]; then
        PROTOBUF_VERSION=3.17.3
        PROTOC_FILENAME=protoc-${PROTOBUF_VERSION}-linux-x86_64.zip
        pushd /home/travis
        wget https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/${PROTOC_FILENAME}
        unzip ${PROTOC_FILENAME}
        bin/protoc --version
        popd
    fi
    exit 0
elif [[ "$#" -ne 0 ]]; then
  die "Unknown argument(s): $*"
fi

misspell -error .

PATH="/home/travis/bin:${PATH}" make proto && \
    git status --porcelain 2>&1 | fail_on_output || \
    (git status; git --no-pager diff; exit 1)

# Perform these checks on each module.
for MOD_FILE in $(find . -name 'go.mod'); do
  MOD_DIR=$(dirname ${MOD_FILE})
  pushd ${MOD_DIR}
  go vet -all ./... | fail_on_output
  gofmt -s -d -l . 2>&1 | fail_on_output
  goimports -l . 2>&1 | not grep -vE "\.pb\.go"
  golint ./... 2>&1 | fail_on_output

  go mod tidy
  git status --porcelain 2>&1 | fail_on_output || \
    (git status; git --no-pager diff; exit 1)
  popd
done

SC_OUT="$(mktemp)"
staticcheck -go 1.17 -checks 'inherit,-ST1015' ./... > "${SC_OUT}" || true
# Error if anything other than deprecation warnings are printed.
not grep -v "is deprecated:.*SA1019" "${SC_OUT}"
