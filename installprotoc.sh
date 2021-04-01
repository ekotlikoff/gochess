#!/usr/bin/env bash

set -ex  # Exit on error; debugging enabled.
set -o pipefail  # Fail a pipe if any sub-command fails.

PROTOBUF_VERSION=3.14.0
PROTOC_FILENAME=protoc-${PROTOBUF_VERSION}-linux-x86_64.zip
pushd /home/travis
wget https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/${PROTOC_FILENAME}
unzip ${PROTOC_FILENAME}
bin/protoc --version
popd
