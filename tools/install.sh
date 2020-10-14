#!/bin/bash

cd `dirname $0`

set -e

cd ..
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH

[[ ! -d "${GOBIN}" ]] && mkdir -p "${GOBIN}"

go install tools/gomodrun.go

export GOBIN=$PWD/bin
export PATH="${GOBIN}:${PATH}"

echo "GOBIN=${GOBIN}"

go install github.com/golang/protobuf/protoc-gen-go/...

[[ -f protoc.zip ]] && rm protoc.zip
[[ -f bin/protoc ]] && rm bin/protoc
[[ -d tmp ]] && rm -rf tmp
mkdir tmp
cd tmp
if [[ "$OSTYPE" == "darwin"* ]]; then
    wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-osx-x86_64.zip
else
    wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip
fi
ls -la protoc.zip
unzip protoc.zip
mv bin/protoc ../bin
cd ..
rm -rf tmp
chmod +x bin/protoc

# https://github.com/golangci/golangci-lint#go Please, do not installDaemon golangci-lint by go get
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN" v1.31.0
