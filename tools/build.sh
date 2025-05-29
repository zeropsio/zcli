#!/bin/bash

export VERSION="$(git rev-parse --abbrev-ref HEAD):$(git describe --tags)-($(git config --get user.name):<$(git config --get user.email)>)"
go build -tags devel \
         -o bin/$1 \
         -gcflags="all=-l -N" \
         -ldflags="all=\"-X=github.com/zeropsio/zcli/src/cmd.version=${VERSION}\"" \
         cmd/zcli/main.go
