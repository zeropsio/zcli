# CLAUDE.md - zcli Developer Guide

## Project Overview

**zcli** is the official Command Line Interface for [Zerops](https://zerops.io/), a cloud platform for deploying and managing applications. Written in Go 1.24, it provides comprehensive functionality for project management, service deployment, VPN connectivity, and log streaming.

**Repository**: `github.com/zeropsio/zcli`

### Supported Platforms
- Linux (amd64, i386)
- macOS (amd64, arm64)
- Windows (amd64)
- NixOS

---

## Quick Reference

### Essential Commands
```bash
# Build
go build -o zcli ./cmd/zcli/main.go

# Test
go test -v ./cmd/... ./src/...

# Lint (multi-platform)
make lint

# Build all platforms
make all

# Run UI showcase
make showcase
```

### Project Structure
```
zcli/
├── cmd/zcli/main.go      # Entry point
├── src/
│   ├── cmd/              # CLI commands (login, project, service, vpn, etc.)
│   ├── cmdBuilder/       # Command construction framework (wraps Cobra)
│   ├── entity/           # Domain models (Project, Service, Org, etc.)
│   ├── uxBlock/          # Terminal UI components (Bubble Tea-based)
│   ├── zeropsRestApiClient/  # API client wrapper
│   ├── cliStorage/       # Local credential/config storage
│   ├── i18n/             # Internationalization (English translations)
│   └── ...               # Supporting packages
├── tools/                # Build scripts and npm package
└── .github/workflows/    # CI/CD pipelines
```

---

## Architecture Deep Dive

### Command Framework (`src/cmdBuilder/`)

The CLI uses a custom wrapper around [Cobra](https://github.com/spf13/cobra) with a fluent builder pattern:

```go
func myCmd() *cmdBuilder.Cmd {
    return cmdBuilder.NewCmd().
        Use("mycommand").
        Short(i18n.T(i18n.CmdDescMyCommand)).
        ScopeLevel(cmdBuilder.ScopeProject()).  // Requires project context
        Arg("argName", cmdBuilder.OptionalArg()).
        StringFlag("flag-name", "default", i18n.T(i18n.FlagDesc)).
        BoolFlag("verbose", false, i18n.T(i18n.VerboseFlag)).
        HelpFlag(i18n.T(i18n.CmdHelpMyCommand)).
        LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
            // Implementation for authenticated users
            return nil
        }).
        GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
            // Implementation for unauthenticated users
            return nil
        })
}
```

**Key Types**:
- `LoggedUserCmdData`: Contains `RestApiClient`, `CliStorage`, `UxBlocks`, `Project`, `Service`, `Params`, `Args`
- `GuestCmdData`: Limited context for unauthenticated commands
- `ScopeLevel`: Interface for project/service scope resolution

### Scope System

Commands can require different scope levels:
- `cmdBuilder.ScopeProject()` - Requires project selection
- `cmdBuilder.ScopeService()` - Requires service selection within a project
- Options: `WithCreateNewProject()`, `WithCreateNewService()`, `WithSkipSelectProject()`

The scope system handles:
1. Persisted scope from `zcli scope project`
2. CLI flags (`--project-id`, `--service-id`)
3. Positional arguments
4. Interactive selection (when in terminal mode)

### Entity Models (`src/entity/`)

Core domain objects:
```go
type Project struct {
    Id          uuid.ProjectId
    Name        types.String
    Mode        enum.ProjectModeEnum
    OrgId       uuid.ClientId
    OrgName     types.String
    Description types.Text
    Status      enum.ProjectStatusEnum
}

type Service struct {
    Id                  uuid.ServiceStackId
    ProjectId           uuid.ProjectId
    Name                types.String
    Status              enum.ServiceStackStatusEnum
    ServiceTypeId       stringId.ServiceStackTypeId
    ServiceTypeCategory enum.ServiceStackTypeCategoryEnum
}
```

### UX Components (`src/uxBlock/`)

Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss):

- **Spinners**: Async operation progress (`uxBlock.Spinner`)
- **Selectors**: Interactive list selection (`models/selector/`)
- **Prompts**: Yes/No confirmations (`models/prompt/`)
- **Inputs**: Text input with validation (`models/input/`)
- **Tables**: Data display (`models/table/`)
- **Log Views**: Streaming log display (`models/logView/`)

**Usage Pattern**:
```go
err := uxHelpers.ProcessCheckWithSpinner(ctx, cmdData.UxBlocks, []uxHelpers.Process{{
    F:                   myAsyncFunc,
    RunningMessage:      "Processing...",
    SuccessMessage:      "Done!",
    ErrorMessageMessage: "Failed!",
}})
```

### Storage (`src/cliStorage/`)

Local JSON file storage for credentials and state:
```go
type Data struct {
    Token                 string
    RegionData            region.Item
    ScopeProjectId        uuid.ProjectIdNull
    ProjectVpnKeyRegistry map[uuid.ProjectId]entity.VpnKey
}
```

Storage paths vary by OS (see `src/constants/`):
- macOS: `~/Library/Application Support/zerops/` or `~/.zerops/`
- Linux: `~/.config/zerops/` or `~/.zerops/`
- Windows: `%APPDATA%\Zerops\`

### API Client (`src/zeropsRestApiClient/`)

Wraps the `github.com/zeropsio/zerops-go` SDK:
```go
client := zeropsRestApiClient.NewAuthorizedClient(token, regionUrl)
response, err := client.GetUserInfo(ctx)
```

### Internationalization (`src/i18n/`)

All user-facing strings use the translation system:
```go
i18n.T(i18n.LoginSuccess, fullName, email)  // "You are logged as %s <%s>"
```

Constants are defined in `i18n.go`, translations in `en.go`.

---

## Key Workflows

### Push/Deploy Flow

1. Read `zerops.yml` from working directory
2. Validate YAML against service configuration
3. Create app version via API
4. Archive files (respecting `.gitignore` and `.deployignore`)
5. Stream archive to Zerops
6. Trigger build/deploy pipeline
7. Poll process status until completion

### VPN Connection (`vpn up`)

1. Check for existing WireGuard interface
2. Get or create private key (stored per-project)
3. Exchange public key with Zerops API
4. Generate WireGuard config file
5. Execute `wg-quick up`
6. Verify connectivity via ping

---

## Testing

### Running Tests
```bash
go test -v ./cmd/... ./src/...
```

### Test Patterns
- Table-driven tests with `testify/require`
- Mock generation via `github.com/golang/mock/mockgen`
- See `src/uxBlock/mocks/` for mock examples

### Example Test Structure
```go
func TestConvertArgs(t *testing.T) {
    tests := []struct {
        name    string
        args    args
        want    map[string][]string
        wantErr string
    }{
        // test cases...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // assertions
        })
    }
}
```

---

## Build & Release

### Local Development Build
```bash
./tools/build.sh zcli            # Uses git branch/tag for version
GOOS=darwin GOARCH=arm64 ./tools/build.sh zcli.darwin
```

### Production Build (via CI)
```bash
go build \
    -o zcli \
    -ldflags "-s -w -X github.com/zeropsio/zcli/src/version.version=v1.0.0" \
    ./cmd/zcli/main.go
```

### CI/CD Workflows (`.github/workflows/`)

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `main.yml` | Push/PR to main | Build, test, lint (multi-platform) |
| `release.yml` | GitHub release | Build binaries, publish to NPM, notify Discord |
| `pre-release.yml` | Pre-release | Preview builds |

### Release Artifacts
- `zcli-linux-amd64`, `zcli-linux-i386`
- `zcli-darwin-amd64`, `zcli-darwin-arm64`
- `zcli-win-x64.exe`
- NPM package: `@zerops/zcli`

---

## Linting

Uses `golangci-lint` v1.64.7 with extensive ruleset (see `.golangci.yaml`):

```bash
# Install
./tools/install.sh

# Run
gomodrun golangci-lint run ./cmd/... ./src/... --verbose
```

Key enabled linters: `gosec`, `govet`, `errcheck`, `staticcheck`, `gocritic`, `gosimple`, `ineffassign`, `unused`, and 60+ more.

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ZEROPS_TOKEN` | Authentication token (takes precedence over stored login) |
| `ZEROPS_TERMINAL_MODE` | Terminal mode: `auto`, `enabled`, `disabled` |
| `ZEROPS_LOG_FILE_PATH` | Custom log file path |
| `ZEROPS_DATA_FILE_PATH` | Custom data file path |
| `ZEROPS_WG_CONFIG_PATH` | Custom WireGuard config path |
| `ZEROPS_VERSIONNAME` | Custom version name for deployments |

---

## Generic Utilities (`src/gn/`)

Reusable generic functions used throughout:

```go
gn.Must(value, err)                    // Panic on error
gn.Ptr(value)                          // Get pointer to value
gn.FilterSlice(slice, predicate)       // Filter slice
gn.TransformSlice(slice, transform)    // Map slice
gn.FindFirst(slice, predicate)         // Find first match
gn.ApplyOptions(options...)            // Functional options pattern
gn.MergeMaps(maps...)                  // Merge multiple maps
gn.IsOneOf(val, values...)             // Check membership
```

---

## Adding New Commands

1. Create `src/cmd/myCommand.go`:
```go
package cmd

func myCommandCmd() *cmdBuilder.Cmd {
    return cmdBuilder.NewCmd().
        Use("my-command").
        Short(i18n.T(i18n.CmdDescMyCommand)).
        // ... configure flags, args, scope
        LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
            // Implementation
            return nil
        })
}
```

2. Register in parent command (e.g., `src/cmd/root.go`):
```go
AddChildrenCmd(myCommandCmd())
```

3. Add i18n strings to `src/i18n/i18n.go` and `src/i18n/en.go`

4. Write tests in `src/cmd/myCommand_test.go`

---

## Error Handling

### User Errors
```go
return errorsx.NewUserError("message", originalErr)
```

### API Error Conversion
```go
return errorsx.Convert(err,
    errorsx.ErrorCode(errorCode.ProjectNotFound),
    errorsx.InvalidUserInput("fieldName"),
)
```

### Error Display
Errors are automatically formatted via `cmdBuilder.printError()`:
- User errors: Display message only
- API errors: Display message + meta in YAML
- Ctrl+C: Display "canceled" info

---

## Dependencies

### Core
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `github.com/zeropsio/zerops-go` - Zerops API SDK

### UI
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - UI components

### VPN
- `golang.zx2c4.com/wireguard/wgctrl` - WireGuard control

### Utilities
- `github.com/pkg/errors` - Error wrapping
- `github.com/google/uuid` - UUIDs
- `github.com/gorilla/websocket` - WebSocket (logs)
- `gopkg.in/yaml.v3` - YAML parsing

---

## Common Patterns

### Checking Terminal Mode
```go
if !terminal.IsTerminal() {
    return errors.New("Interactive selection requires terminal")
}
```

### Process Monitoring
```go
uxHelpers.CheckZeropsProcess(processId, cmdData.RestApiClient)
```

### Interactive Selectors
```go
project, selected, err := cmdData.ProjectSelector(ctx, cmdData)
service, err := uxHelpers.PrintServiceSelector(ctx, restApiClient, projectId)
```

### Flags with Shorthand
```go
StringFlag("project-id", "", desc, cmdBuilder.ShortHand("P"))
BoolFlag("verbose", false, desc, cmdBuilder.ShortHand("v"))
```

---

## Debugging

### View Debug Logs
```bash
zcli status show-debug-logs
```

### Log File Locations
- macOS: `/usr/local/var/log/zerops.log` or `~/.zerops/zerops.log`
- Linux: `/var/log/zerops.log` or `~/.zerops/zerops.log`
- Windows: `%APPDATA%\Zerops\zerops.log`

### Verbose Mode
```bash
zcli push --verbose
```

---

## External Resources

- **Documentation**: https://docs.zerops.io/references/cli
- **Discord**: https://discord.com/invite/WDvCZ54
- **Support**: https://support.zerops.io
- **API SDK**: https://github.com/zeropsio/zerops-go
