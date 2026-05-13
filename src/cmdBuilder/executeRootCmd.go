package cmdBuilder

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/logfile"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/apiError"
)

// RunOptions controls how RunRootCmd executes. The zero value matches the
// production defaults used by ExecuteRootCmd.
type RunOptions struct {
	// Ctx is the root context. If nil, a fresh context.Background() is used
	// and OS signals are wired to cancel it.
	//nolint:containedctx // intentional: callers (tests, ExecuteRootCmd) pass the root ctx by value
	Ctx context.Context
	// Args overrides os.Args[1:] when non-nil. Useful for tests.
	Args []string
	// Stdout receives command output. Defaults to os.Stdout when nil.
	Stdout io.Writer
	// Stderr receives error/log output (including uxBlock messages).
	// Defaults to os.Stderr when nil.
	Stderr io.Writer
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
	os.Exit(RunRootCmd(rootCmd, RunOptions{}))
}

// RunRootCmd is the test-friendly entry point. It never calls os.Exit and
// instead returns the exit code the process should use.
func RunRootCmd(rootCmd *Cmd, opts RunOptions) int {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	ctx := opts.Ctx
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithCancel(context.Background())
		regSignals(cancel)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()
	ctx = support.Context(ctx)

	isTerminal := terminal.IsTerminal()
	terminalWidth, terminalHeight, _ := term.GetSize(0)

	installLogger(verboseFromArgs(os.Args[1:]))

	uxBlocks := uxBlock.NewBlocks(isTerminal, terminalWidth, terminalHeight, cancel)

	cliStorage, err := createCliStorage()
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	flagParams := flagParams.New(stderr)

	cobraCmd, err := buildCobraCmd(rootCmd, flagParams, uxBlocks, cliStorage, stdout, stderr)
	if err != nil {
		return errorExitCode(err, uxBlocks)
	}

	cobraCmd.SetGlobalNormalizationFunc(normalizeFlagNames)
	cobraCmd.SetOut(stdout)
	cobraCmd.SetErr(stderr)
	if opts.Args != nil {
		cobraCmd.SetArgs(opts.Args)
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
	slog.Debug("command.error", "err", fmt.Sprintf("%+v", err))

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

// installLogger assembles slog handlers (file + optional stderr) and
// installs the resulting logger as slog.Default.
//
// The file sink is always built at LevelDebug so the log file is a
// complete trace regardless of the --verbose flag; --verbose only
// controls whether the stderr mirror is attached (and at which level).
func installLogger(verbose bool) {
	stderrLevel := slog.LevelInfo
	if verbose {
		stderrLevel = slog.LevelDebug
	}

	var handlers []slog.Handler
	if logFilePath, fileMode, err := constants.LogFilePath(); err != nil {
		// Best effort: warn directly to stderr; logger isn't up yet.
		fmt.Fprintln(os.Stderr, styles.WarningLine(err.Error()).String())
	} else {
		handlers = append(handlers, logfile.New(logfile.Config{
			Path: logFilePath,
			Mode: fileMode,
			Opts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}))
	}
	if verbose {
		handlers = append(handlers, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: stderrLevel}))
	}

	var h slog.Handler
	switch len(handlers) {
	case 0:
		h = slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})
	case 1:
		h = handlers[0]
	default:
		h = slog.NewMultiHandler(handlers...)
	}
	slog.SetDefault(slog.New(h))
}

// verboseFromArgs scans the raw command-line for the persistent --verbose
// (a.k.a. -v) flag so we can configure the logger before cobra parses
// anything.
func verboseFromArgs(args []string) bool {
	for _, a := range args {
		switch a {
		case "-v", "--verbose":
			return true
		}
		if strings.HasPrefix(a, "--verbose=") {
			return a[len("--verbose="):] == "true" || a[len("--verbose="):] == "1"
		}
	}
	return false
}

func regSignals(contextCancel func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		contextCancel()
	}()
}

func createCliStorage() (*cliStorage.Handler, error) {
	filePath, fileMode, err := constants.CliDataFilePath()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[cliStorage.Data](
		storage.Config{
			FilePath: filePath,
			FileMode: fileMode,
		},
	)
	return &cliStorage.Handler{Handler: s}, err
}
