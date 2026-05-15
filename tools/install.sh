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
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN" v2.1.6

GORELEASER_VERSION=v2.5.0
GORELEASER_OS=$(uname -s)
GORELEASER_ARCH=$(uname -m)
curl -sSfL "https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/goreleaser_${GORELEASER_OS}_${GORELEASER_ARCH}.tar.gz" \
  | tar -xz -C "$GOBIN" goreleaser
chmod +x "$GOBIN/goreleaser"
