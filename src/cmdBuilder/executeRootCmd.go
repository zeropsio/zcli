package cmdBuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/apiError"
)

// runOptions controls how RunRootCmd executes. Build via the With... helpers.
type runOptions struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
}

// RunOption configures a RunRootCmd invocation.
type RunOption func(*runOptions)

// WithArgs overrides os.Args[1:]. Useful for tests.
func WithArgs(args []string) RunOption {
	return func(o *runOptions) { o.args = args }
}

// WithStdout redirects command output. Defaults to os.Stdout.
func WithStdout(w io.Writer) RunOption {
	return func(o *runOptions) { o.stdout = w }
}

// WithStderr redirects error/log output (including uxBlock messages).
// Defaults to os.Stderr.
func WithStderr(w io.Writer) RunOption {
	return func(o *runOptions) { o.stderr = w }
}

func newRunOptions(opts ...RunOption) runOptions {
	o := runOptions{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

var matchFirstCap = regexp.MustCompile("([A-Z]+)")

func camelCaseToKebabCase(camel string) string {
	return strings.ToLower(matchFirstCap.ReplaceAllString(camel, "-${1}"))
}

func normalizeFlagNames(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(camelCaseToKebabCase(name))
}

// ExecuteRootCmd runs the CLI with production defaults and exits the process
// with the resulting status code.
func ExecuteRootCmd(rootCmd *Cmd) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	code := RunRootCmd(ctx, rootCmd)
	stop()
	os.Exit(code)
}

// RunRootCmd is the test-friendly entry point. It never calls os.Exit and
// instead returns the exit code the process should use.
func RunRootCmd(ctx context.Context, rootCmd *Cmd, opts ...RunOption) int {
	o := newRunOptions(opts...)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = support.Context(ctx)

	isTerminal := terminal.IsTerminal()
	terminalWidth, terminalHeight, _ := term.GetSize(0)
	outputLogger, debugFileLogger := createLoggers(isTerminal, o.stderr)

	uxBlocks := uxBlock.NewBlocks(outputLogger, debugFileLogger, isTerminal, terminalWidth, terminalHeight, cancel)

	profileManager, err := createProfile(uxBlocks)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	flagParams := flagParams.New(o.stderr)

	//nolint:contextcheck // cobra threads ctx into the run-func via ExecuteContext below; pulling from cobraCmd.Context() inside createCmdRunFunc is the canonical pattern.
	cobraCmd, err := buildCobraCmd(rootCmd, flagParams, uxBlocks, profileManager, o.stdout, o.stderr)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	cobraCmd.SetGlobalNormalizationFunc(normalizeFlagNames)
	cobraCmd.SetOut(o.stdout)
	cobraCmd.SetErr(o.stderr)
	if o.args != nil {
		cobraCmd.SetArgs(o.args)
	}

	if err := cobraCmd.ExecuteContext(ctx); err != nil {
		return errorExitCode(err, uxBlocks)
	}
	return 0
}

func errorExitCode(err error, uxBlocks uxBlock.UxBlocks) int {
	if err == nil {
		return 0
	}
	uxBlocks.LogDebug(fmt.Sprintf("error: %+v", err))

	// A command that manages its own output and needs a specific exit status
	// (e.g. `upgrade --check`) signals it with an ExitError; return the code
	// as-is without printing anything further.
	if exitErr := errorsx.AsExitError(err); exitErr != nil {
		return exitErr.Code
	}

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintErrorText(err.Error())
		return 1
	}

	var apiErr apiError.Error
	if errors.As(err, &apiErr) {
		uxBlocks.PrintErrorText(apiErr.GetMessage())
		if apiErr.GetMeta() != nil {
			meta, err := yaml.Marshal(apiErr.GetMeta())
			if err != nil {
				uxBlocks.PrintErrorText(fmt.Sprintf("couldn't parse meta of error: %s", apiErr.GetMessage()))
			}
			uxBlocks.PrintErrorText(string(meta))
		}
		return 1
	}

	if errors.Is(err, models.ErrCtrlC) {
		uxBlocks.PrintInfo(styles.InfoLine("canceled"))
		return 0
	}

	uxBlocks.PrintErrorText(err.Error())
	return 1
}

func createLoggers(isTerminal bool, stderr io.Writer) (*logger.Handler, *logger.Handler) {
	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
		Out:        stderr,
	})

	loggerFilePath, fileMode, err := constants.LogFilePath()
	if err != nil {
		outputLogger.Warning(styles.WarningLine(err.Error()))
	}

	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: loggerFilePath,
		FileMode: fileMode,
	})

	return outputLogger, debugFileLogger
}

func createProfile(uxBlocks *uxBlock.Blocks) (*profile.Manager, error) {
	filePath, fileMode, err := constants.CliProfileFilePath()
	if err != nil {
		return nil, err
	}

	legacyPath, _, err := constants.CliLegacyDataFilePath()
	if err != nil {
		return nil, err
	}

	secretsPath := os.Getenv(constants.CliSecretsFilePathEnvVar)
	if secretsPath == "" {
		secretsPath = filepath.Join(filepath.Dir(filePath), constants.CliSecretsFileName)
	}

	mode := secrets.Mode(os.Getenv(constants.SecretsBackendEnvVar))
	if mode == "" {
		mode = peekStoredSecretsMode(filePath)
	}

	secretsBackend, err := secrets.New(secrets.Config{
		Mode:     mode,
		Service:  "zerops-cli",
		FilePath: secretsPath,
		FileMode: 0o600,
		Log:      func(msg string) { uxBlocks.PrintWarningText(msg) },
	})
	if err != nil {
		return nil, err
	}

	return profile.NewManager(profile.Config{
		GlobalPath:       filePath,
		GlobalMode:       fileMode,
		LegacyGlobalPath: legacyPath,
		Cwd:              os.Getenv(constants.CliCwdEnvVar), // empty = Manager falls back to os.Getwd()
		Secrets:          secretsBackend,
		Log:              func(msg string) { uxBlocks.PrintInfoText(msg) },
	})
}

// peekStoredSecretsMode reads only the secrets_backend field from the profile JSON, without
// triggering a full load or migration. Lets us resolve the user's preferred backend before
// constructing the secrets layer that the full load needs.
func peekStoredSecretsMode(profilePath string) secrets.Mode {
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		return secrets.ModeAuto
	}
	var probe struct {
		SecretsBackend secrets.Mode `json:"secrets_backend,omitempty"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || probe.SecretsBackend == "" {
		return secrets.ModeAuto
	}
	return probe.SecretsBackend
}
