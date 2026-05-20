.DEFAULT_GOAL := help

BIN := $(CURDIR)/bin
GOLANGCI_LINT_VERSION := v2.12.0
GORELEASER_VERSION := v2.15.4
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
# golangci-lint releases use lowercase OS and amd64/arm64 in tarball names.
GCI_OS := $(shell echo $(UNAME_S) | tr A-Z a-z)
GCI_ARCH := $(if $(filter $(UNAME_M),x86_64),amd64,$(if $(filter $(UNAME_M),aarch64),arm64,$(UNAME_M)))
GCI_DIR := golangci-lint-$(GOLANGCI_LINT_VERSION:v%=%)-$(GCI_OS)-$(GCI_ARCH)

.PHONY: help test test-integration lint all build-dev install install-dev windows-amd linux-amd darwin-amd darwin-arm showcase goreleaser-check goreleaser-snapshot tools clean-tools

# `make install` matches install.sh — puts zcli into ~/.local/bin so it
# lands on the same PATH entry end users get from the install script.
# `make install-dev` uses Go's convention ($GOBIN, else $GOPATH/bin) so
# zcli-dev sits with the user's other Go dev tools.
PROD_INSTALL_DIR := $(HOME)/.local/bin
DEV_INSTALL_DIR  := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)

# Dev-build version metadata (matches what tools/build.sh used to compose).
DEV_VERSION := $(shell git rev-parse --abbrev-ref HEAD):$(shell git describe --tags 2>/dev/null)-($(shell git config --get user.name):<$(shell git config --get user.email)>)
# -gcflags disables inlining and optimizations so the binary is dlv-friendly.
DEV_BUILD := go build \
	-gcflags='all=-l -N' \
	-ldflags='-X "github.com/zeropsio/zcli/src/version.version=$(DEV_VERSION)"'

# Production build flags mirror what .goreleaser.yaml uses for release builds:
# optimized, stripped, version from `git describe`, paths trimmed.
PROD_VERSION := $(shell git describe --tags 2>/dev/null)
PROD_BUILD := go build -trimpath \
	-ldflags='-s -w -X github.com/zeropsio/zcli/src/version.version=$(PROD_VERSION)'

# Self-documenting help. Targets are listed in the order they appear here;
# their description is the text after the '##' on the recipe line.
help: ## Show this help.
	@awk 'BEGIN { FS = ":.*## " } \
	     /^##@ / { printf "\n%s\n", substr($$0, 5); next } \
	     /^[a-zA-Z0-9_-]+:.*## / { printf "  %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Development

test: ## Run the full Go test suite.
	go test -v ./cmd/... ./src/...

test-integration: ## Run just the in-process integration tests (src/cmd).
	go test -v ./src/cmd/...

# Lint each target GOOS in turn so platform-specific build tags get covered.
lint: $(BIN)/.golangci-lint-$(GOLANGCI_LINT_VERSION) ## Run golangci-lint for darwin/arm64, linux/amd64, windows/amd64.
	GOOS=darwin  GOARCH=arm64 $(BIN)/golangci-lint run ./cmd/... ./src/... --verbose
	GOOS=linux   GOARCH=amd64 $(BIN)/golangci-lint run ./cmd/... ./src/... --verbose
	GOOS=windows GOARCH=amd64 $(BIN)/golangci-lint run ./cmd/... ./src/... --verbose

##@ Build

build-dev: ## Build a dev binary for the host into ./bin/zcli (no optimizations, dlv-friendly).
	$(DEV_BUILD) -o $(BIN)/zcli ./cmd/zcli

all: windows-amd linux-amd darwin-amd darwin-arm ## Cross-build all dev targets.

windows-amd: ## Build the windows/amd64 dev binary.
	GOOS=windows GOARCH=amd64 $(DEV_BUILD) -o $(BIN)/zcli.win.exe ./cmd/zcli

linux-amd: ## Build the linux/amd64 dev binary.
	GOOS=linux GOARCH=amd64 $(DEV_BUILD) -o $(BIN)/zcli.linux ./cmd/zcli

darwin-amd: ## Build the darwin/amd64 dev binary.
	GOOS=darwin GOARCH=amd64 $(DEV_BUILD) -o $(BIN)/zcli.darwin.amd64 ./cmd/zcli

darwin-arm: ## Build the darwin/arm64 dev binary.
	GOOS=darwin GOARCH=arm64 $(DEV_BUILD) -o $(BIN)/zcli.darwin.arm64 ./cmd/zcli

##@ Install

install: ## Build a production zcli (stripped, optimized) and install it into ~/.local/bin (matches install.sh).
	@mkdir -p $(PROD_INSTALL_DIR)
	$(PROD_BUILD) -o $(PROD_INSTALL_DIR)/zcli ./cmd/zcli
	@echo "installed $(PROD_INSTALL_DIR)/zcli ($(PROD_VERSION))"

install-dev: ## Build a dev zcli-dev (debug-friendly) and install it into $GOBIN.
	$(DEV_BUILD) -o $(DEV_INSTALL_DIR)/zcli-dev ./cmd/zcli
	@echo "installed $(DEV_INSTALL_DIR)/zcli-dev"

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
	curl -sSfL "https://github.com/golangci/golangci-lint/releases/download/$(GOLANGCI_LINT_VERSION)/$(GCI_DIR).tar.gz" \
		| tar -xz -C $(BIN) --strip-components=1 $(GCI_DIR)/golangci-lint
	chmod +x $(BIN)/golangci-lint
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
