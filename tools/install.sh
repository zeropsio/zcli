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

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

[[ -f protoc.zip ]] && rm protoc.zip
[[ -f bin/protoc ]] && rm bin/protoc
[[ -d include ]] && rm -rf include
[[ -d tmp ]] && rm -rf tmp
mkdir include
mkdir tmp
cd tmp

if [[ "$OSTYPE" == "darwin"* ]]; then
    wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v21.3/protoc-21.3-osx-x86_64.zip
else
    wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v21.3/protoc-21.3-linux-x86_64.zip
fi

ls -la protoc.zip
unzip protoc.zip

mv bin/protoc ../bin
mv -v ./include/* ../include/google

cd ..
chmod +x bin/protoc

rm -rf tmp

# https://github.com/golangci/golangci-lint#go Please, do not installDaemon golangci-lint by go get
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN" v1.55.2
