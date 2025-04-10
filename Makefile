.PHONY: help

define helpMessage
possible targets:
- test
- lint
- all
- windows-amd
- linux-amd
- dawrin-amd
- darwin-arm
- showcase
endef
export helpMessage

help:
	@echo "$$helpMessage"

test:
	 go test -v ./cmd/... ./src/...

lint:
	GOOS=darwin GOARCH=arm64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose
	GOOS=linux GOARCH=amd64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose
	GOOS=windows GOARCH=amd64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose

#########
# BUILD #
#########
all: windows-amd linux-amd darwin-amd darwin-arm

windows-amd:
	GOOS=windows GOARCH=amd64 tools/build.sh zcli.exe

linux-amd:
	GOOS=linux GOARCH=amd64 tools/build.sh zcli.linux

darwin-amd:
	GOOS=darwin GOARCH=amd64 tools/build.sh zcli.darwin.amd64

darwin-arm:
	GOOS=darwin GOARCH=arm64 tools/build.sh zcli.darwin.arm64

#########
# OTHER #
#########

# showcase of ui elements
showcase:
	go run src/uxBlock/showcase/main.go
