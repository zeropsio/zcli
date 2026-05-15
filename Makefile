.DEFAULT_GOAL := help

.PHONY: help test test-integration lint all windows-amd linux-amd darwin-amd darwin-arm showcase goreleaser-check goreleaser-snapshot

# Self-documenting help. Targets are listed in the order they appear here;
# their description is the text after the '##' on the recipe line.
help: ## Show this help.
	@awk 'BEGIN { FS = ":.*## " } \
	     /^##@ / { printf "\n%s\n", substr($$0, 5); next } \
	     /^[a-zA-Z0-9_-]+:.*## / { printf "  %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Development

test: ## Run the full Go test suite.
	go test -v ./cmd/... ./src/...

test-integration: ## Run the integration test suite (devel build tag).
	go test -v -tags devel ./src/cmd/...

# Lint each target GOOS in turn so platform-specific build tags get covered.
lint: ## Run golangci-lint for darwin/arm64, linux/amd64, windows/amd64.
	GOOS=darwin  GOARCH=arm64 gomodrun golangci-lint run ./cmd/... ./src/... --verbose
	GOOS=linux   GOARCH=amd64 gomodrun golangci-lint run ./cmd/... ./src/... --verbose
	GOOS=windows GOARCH=amd64 gomodrun golangci-lint run ./cmd/... ./src/... --verbose

##@ Build

# tools/build.sh embeds version metadata via -ldflags and writes to ./bin/.
all: windows-amd linux-amd darwin-amd darwin-arm ## Cross-build all release targets.

windows-amd: ## Build the windows/amd64 binary.
	GOOS=windows GOARCH=amd64 tools/build.sh zcli.win.exe

linux-amd: ## Build the linux/amd64 binary.
	GOOS=linux GOARCH=amd64 tools/build.sh zcli.linux

darwin-amd: ## Build the darwin/amd64 binary.
	GOOS=darwin GOARCH=amd64 tools/build.sh zcli.darwin.amd64

darwin-arm: ## Build the darwin/arm64 binary.
	GOOS=darwin GOARCH=arm64 tools/build.sh zcli.darwin.arm64

##@ Release tooling

# Both targets shell out to the pinned goreleaser in ./bin via gomodrun;
# tools/install.sh provisions it alongside golangci-lint.
goreleaser-check: ## Validate .goreleaser.yaml without building.
	gomodrun goreleaser check

goreleaser-snapshot: ## Dry-run a full release build to ./dist (no upload, no publish).
	gomodrun goreleaser release --snapshot --clean --skip=publish

##@ Other

showcase: ## Preview uxBlock UI elements in the terminal.
	go run src/uxBlock/showcase/main.go
