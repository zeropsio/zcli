.PHONY: help install

define helpMessage
possible targets:
- test
- test-integration
- lint
- all
- install
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

test-integration:
	 go test -v -tags devel ./src/cmd/...

lint:
	GOOS=darwin GOARCH=arm64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose
	GOOS=linux GOARCH=amd64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose
	GOOS=windows GOARCH=amd64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose

#########
# BUILD #
#########
all: windows-amd linux-amd darwin-amd darwin-arm

windows-amd:
	GOOS=windows GOARCH=amd64 tools/build.sh zcli.win.exe

linux-amd:
	GOOS=linux GOARCH=amd64 tools/build.sh zcli.linux

darwin-amd:
	GOOS=darwin GOARCH=amd64 tools/build.sh zcli.darwin.amd64

darwin-arm:
	GOOS=darwin GOARCH=arm64 tools/build.sh zcli.darwin.arm64

#########
# OTHER #
#########

# build zcli for the host platform and install into $GOPATH/bin
install:
	@VERSION="$$(git rev-parse --abbrev-ref HEAD):$$(git describe --tags)-($$(git config --get user.name):<$$(git config --get user.email)>)" && \
	go install -tags devel \
		-ldflags "-X 'github.com/zeropsio/zcli/src/version.version=$$VERSION'" \
		./cmd/zcli

# showcase of ui elements
showcase:
	go run src/uxBlock/showcase/main.go
