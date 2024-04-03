## help             show this help
.PHONY: help

help:
	@printf "possible values: test, lint"

test:
	 go test -v ./cmd/... ./src/...

lint:
	GOOS=darwin GOARCH=arm64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose
	GOOS=linux GOARCH=amd64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose
	GOOS=windows GOARCH=amd64 gomodrun golangci-lint run  ./cmd/... ./src/... --verbose

build-for-windows-amd:
	 GOOS=windows GOARCH=amd64 go build -o bin/zcli.exe cmd/zcli/main.go

build-for-linux-amd:
	 GOOS=linux GOARCH=amd64 go build -o bin/zcli cmd/zcli/main.go

build-for-darwin-arm:
	 GOOS=darwin GOARCH=arm64 go build -o bin/zcli cmd/zcli/main.go
