## help             show this help
.PHONY: help

define helpMessage
possible values:
	test
 	lint
 	windows-amd
 	linux-amd
 	darwin-arm
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

windows-amd:
	 GOOS=windows GOARCH=amd64 go build -o bin/zcli.exe cmd/zcli/main.go

linux-amd:
	 GOOS=linux GOARCH=amd64 go build -o bin/zcli.linux cmd/zcli/main.go

darwin-arm:
	 GOOS=darwin GOARCH=arm64 go build -o bin/zcli.darwin cmd/zcli/main.go
