# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`zcli` is the Zerops CLI written in Go (module `github.com/zeropsio/zcli`, Go 1.26). The entrypoint is `cmd/zcli/main.go`, which delegates to `src/cmd.ExecuteCmd()`. All implementation lives under `src/`.

## Common commands

- `make test` — runs `go test -v ./cmd/... ./src/...`. Run a single test with `go test -v -run TestName ./src/<pkg>/...`.
- `make lint` — runs the pinned `./bin/golangci-lint` for darwin/arm64, linux/amd64, and windows/amd64. Config: `.golangci.yaml`.
- `make build-dev` — build a dev binary for the host into `./bin/zcli` (build tag `devel`, debug-friendly via `-gcflags='all=-l -N'`, version injected from git).
- `make all` — cross-builds dev binaries for windows-amd, linux-amd, darwin-amd, darwin-arm. Single-target builds: `make linux-amd` etc. All build targets share the `DEV_BUILD` recipe in the Makefile (version metadata, devel tag, gcflags).
- `make goreleaser-check` / `make goreleaser-snapshot` — validate `.goreleaser.yaml` and dry-run a release build to `./dist`.
- `make tools` — install pinned dev tooling (`golangci-lint`, `goreleaser`) into `./bin`. Versions are declared at the top of the Makefile; bumping them triggers reinstall via stamp files. `make lint` and the goreleaser targets depend on `tools`, so first invocation auto-installs.
- `make showcase` — runs `src/uxBlock/showcase/main.go` to preview UI elements.

## Architecture

### Command layer (`src/cmd`, `src/cmdBuilder`)

Commands are not declared directly against Cobra. Instead, `src/cmdBuilder` wraps Cobra with a fluent builder (`cmdBuilder.NewCmd().Use(...).Arg(...).StringFlag(...).LoggedUserRunFunc(...)`) and `buildCobraCmd.go` translates it into a `*cobra.Command`. The root command lives in `src/cmd/root.go` and registers all subcommands via `AddChildrenCmd`.

Two run-function flavors are central to how commands behave:

- `GuestRunFunc(func(ctx, *GuestCmdData) error)` — runs when there is no auth token. Has access to `CliStorage`, `UxBlocks`, parsed `Args`/`Params`, `Stdout`/`Stderr` printers.
- `LoggedUserRunFunc(func(ctx, *LoggedUserCmdData) error)` — runs when a token is present (from `cliStorage` or the `CliTokenEnvVar` env var). Embeds `*GuestCmdData` and adds `RestApiClient` (a `zeropsRestApiClient.Handler` pointed at the configured region/host), plus optional resolved `Project`/`Service` entities and a `ProjectSelector`.

If only `GuestRunFunc` is defined, it runs for logged-in users too. If only `LoggedUserRunFunc` is defined and no token is set, the command exits with `UnauthenticatedUser`. Implementation: `src/cmdBuilder/createRunFunc.go`.

### Scope levels (`src/cmdBuilder/scopeProject.go`, `scopeService.go`)

Commands that operate on a project or service set `.ScopeLevel(cmdBuilder.ScopeProject(...))` or `.ScopeLevel(cmdBuilder.ScopeService(...))`. The scope level adds the relevant flags/args and, before the run function executes, resolves the target by ID/name into `LoggedUserCmdData.Project` / `.Service` — possibly creating a new project/service interactively when `WithCreateNewProject()` / `WithCreateNewService()` are passed (see `serviceDeploy.go` for an example). Persisted "scope" (saved project) is read from `CliStorage.Data().ScopeProjectId`.

### Flag normalization

`ExecuteRootCmd` sets a Cobra normalization function that converts camelCase flag names to kebab-case (`src/cmdBuilder/executeRootCmd.go`). Flags can be defined as camelCase in code and used as kebab-case on the CLI.

### Persistent state and config

- `cliStorage` (`src/cliStorage`) wraps `storage.New[Data]` (`src/storage`) — a JSON-backed file used to persist the login token, region selection, saved project scope, and VPN key registry. Path comes from `constants.CliDataFilePath()`.
- `constants` exposes platform-specific paths, env var names (e.g. `CliTokenEnvVar`), and the default region (used when no region is configured).
- Log file path comes from `constants.LogFilePath()`; loggers (`src/logger`) split human output from a debug file logger that captures `LogDebug`/error traces.

### UX layer (`src/uxBlock`, `src/uxHelpers`, `src/printer`, `src/terminal`)

All terminal output and interactive prompts go through `uxBlock.Blocks` (constructed in `executeRootCmd.go`). It owns the loggers, terminal size, an `isTerminal` flag, and a cancel function wired to SIGINT/SIGTERM. `uxBlock/models` provides Bubble Tea TUI models (selectors, spinners, tables); `uxHelpers` wraps common flows like `ProjectSelector`. `printer` is the lower-level line printer used directly by commands (`cmdData.Stdout.PrintLines(...)`). `styles` centralizes lipgloss styles, including the custom Cobra help template defined in `src/cmd/root.go`.

### API client and entities

- `src/zeropsRestApiClient` constructs an authorized client around `github.com/zeropsio/zerops-go`. The host comes from the user's saved `RegionData.Address` or `constants.DefaultRegion`, prefixed with `https://`.
- `src/entity` defines domain types (`Project`, `Service`, `AppVersion`, `Container`, `Process`, `Location`, `Org`, `VpnKey`). `src/entity/repository` contains query helpers (`GetProjectById`, etc.) layered on the REST client.
- Errors from the API are unwrapped in `executeRootCmd.printError`: user-facing `errorsx.UserError`, structured `apiError.Error` (printed with YAML-formatted metadata), and Ctrl-C cancellations are each handled distinctly.

### Other notable packages

- `src/archiveClient` — packs source for `service deploy`/`service push`, honoring `.gitignore` and a `deploy-git-folder` flag.
- `src/yamlReader` — reads `zerops.yaml` (or a path passed via `--zerops-yaml-path`).
- `src/wg` + `src/nettools` — WireGuard interface management for `zcli vpn` (`InterfaceExists`, key handling, etc.); requires the `wireguard` and `ping` binaries on the host.
- `src/serviceLogs` — streaming/log retrieval used by `zcli service log`.
- `src/i18n` — message catalog (`en.go`); all user-visible strings use `i18n.T(i18n.<Key>, args...)`. Add new strings to `en.go` rather than inlining literals.
- `src/flagParams` — typed param reader (`Params.GetBool/GetString/...`) bound from Cobra flags inside the run func.
- `src/optional` — `Null[T]` type used by `LoggedUserCmdData.Project`/`Service`.
- `src/gn` — small generic helpers (`TransformSlice`, etc.).

## Adding a new command

1. Create `src/cmd/<name>.go` exporting a `*cmdBuilder.Cmd` constructor.
2. Wire flags/args, choose `ScopeLevel` if it targets a project/service, and provide `LoggedUserRunFunc` (and/or `GuestRunFunc`).
3. Register it in `rootCmd()` in `src/cmd/root.go` (or as a child of an existing parent like `serviceCmd`, `projectCmd`, `vpnCmd`).
4. Add any new user-facing strings to `src/i18n/en.go` and reference them via `i18n.T(...)`.

## Code style

### Multiline function calls

When breaking a function call across lines, do not use the "half line" form (some arguments on the opening line, the rest continued below). Either keep everything on one line, or put `func(` alone on the opening line, every argument on its own line with a trailing comma, and `)` on its own closing line. Printf-style format strings may keep their format args on the same line as the format string — treat that group as a single argument block.

```go
// good — single line
require.Equal(t, 0, res.ExitCode)

// good — fully expanded, each arg on its own line, trailing commas
require.NotEqualf(
    t,
    0,
    res.ExitCode,
    "expected non-zero exit\n--- stderr ---\n%s\n--- stdout ---\n%s", res.Stderr, res.Stdout,
)

// bad — half-line: first args ride on the opening line, rest spill below
require.NotEqualf(t, 0, res.ExitCode,
    "expected non-zero exit: %s", res.Stderr)
```

## Distribution

Released via GitHub Actions (`.github/workflows/`) and republished to npm as `@zerops/zcli` (`tools/npm`). Install scripts: `install.sh` (Linux/macOS), `install.ps1` (Windows). Nix flake (`flake.nix`, `default.nix`) supports `nix develop`/`nix build`.

## Comments

Default: no comment. Add one only when the *why* is non-obvious (constraint, workaround, surprising invariant) or the *what* is genuinely hard to parse from the code (nested data shape, non-trivial algorithm). Lead *why* comments with the motivation, not a "what" preamble.

Write comments as natural prose, wrapping at thought boundaries, never at column boundaries. Do NOT break lines at 80 or 100 columns mid-sentence. A comment line should read like a sentence you'd speak aloud: it ends where the thought ends, not where a column counter fires.

Short related clauses that flow (a consequence, an addendum, a clarification) stay on the same line, even past a period, as long as they're part of the same thought. A new line means a new thought. Don't split one thought across lines AND don't pack unrelated thoughts onto one line. When enumerating parallel items in prose, join them with commas, not periods.

Target line length is 160-175 chars. Lines may extend to 180 chars when cutting shorter would break mid-clause, but 180 is a hard ceiling, not a target. If a thought doesn't fit in 180 chars, split at a natural clause boundary and continue on a new `//` line.

Don't use em-dashes (`—`) anywhere in comments, use a normal dash (`-`). Don't use semicolons (`;`) as thought separators. End the first thought with a period and start a new sentence. Semicolons for tight lists inside a single thought are fine.

```go
// Value caches the first successful fetch result. Errors are not cached, so the next Get retries.
// Fetch is passed at Get time (not at construction), so callers can close over per-call context without storing it on the struct.
```
