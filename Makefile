.DEFAULT_GOAL := help

BIN := $(CURDIR)/bin
GOLANGCI_LINT_VERSION := v2.1.6
GORELEASER_VERSION := v2.5.0
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

.PHONY: help test test-integration lint all windows-amd linux-amd darwin-amd darwin-arm showcase goreleaser-check goreleaser-snapshot tools clean-tools

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
lint: $(BIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION) ## Run golangci-lint for darwin/arm64, linux/amd64, windows/amd64.
	GOOS=darwin  GOARCH=arm64 $(BIN)/golangci-lint run ./cmd/... ./src/... --verbose
	GOOS=linux   GOARCH=amd64 $(BIN)/golangci-lint run ./cmd/... ./src/... --verbose
	GOOS=windows GOARCH=amd64 $(BIN)/golangci-lint run ./cmd/... ./src/... --verbose

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

goreleaser-check: $(BIN)/.goreleaser-$(GORELEASER_VERSION) ## Validate .goreleaser.yaml without building.
	$(BIN)/goreleaser check

goreleaser-snapshot: $(BIN)/.goreleaser-$(GORELEASER_VERSION) ## Dry-run a full release build to ./dist (no upload, no publish).
	$(BIN)/goreleaser release --snapshot --clean --skip=publish

##@ Tooling

tools: $(BIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION) $(BIN)/.goreleaser-$(GORELEASER_VERSION) ## Install pinned dev tooling into ./bin.

clean-tools: ## Remove installed tooling from ./bin.
	rm -f $(BIN)/golangci-lint $(BIN)/goreleaser $(BIN)/.golangci-lint-* $(BIN)/.goreleaser-*

# Stamp files encode the pinned version. Bumping a version above retargets
# the dependency, the old stamp is removed inside the recipe, and the tool
# gets reinstalled on the next `make lint` / `make goreleaser-*` / `make tools`.
$(BIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION):
	@mkdir -p $(BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $(BIN) $(GOLANGCI_LINT_VERSION)
	@rm -f $(BIN)/.golangci-lint-*
	@touch $@

$(BIN)/.goreleaser-$(GORELEASER_VERSION):
	@mkdir -p $(BIN)
	curl -sSfL "https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(UNAME_S)_$(UNAME_M).tar.gz" \
		| tar -xz -C $(BIN) goreleaser
	chmod +x $(BIN)/goreleaser
	@rm -f $(BIN)/.goreleaser-*
	@touch $@

##@ Other

showcase: ## Preview uxBlock UI elements in the terminal.
	go run src/uxBlock/showcase/main.go
