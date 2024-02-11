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

rm -rf tmp

# https://github.com/golangci/golangci-lint#go Please, do not installDaemon golangci-lint by go get
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN" v1.55.2
