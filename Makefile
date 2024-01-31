## help             show this help
.PHONY: help

help:
	@printf "possible values: test, lint"

test:
	go test -v ./cmd/... ./src/...

lint:
	gomodrun golangci-lint run  ./cmd/... ./src/... --verbose