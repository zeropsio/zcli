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

var matchFirstCap = regexp.MustCompile("([A-Z]+)")

func camelCaseToKebabCase(camel string) string {
	return strings.ToLower(matchFirstCap.ReplaceAllString(camel, "-${1}"))
}

func normalizeFlagNames(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(camelCaseToKebabCase(name))
}

func ExecuteRootCmd(rootCmd *Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	regSignals(cancel)
	ctx = support.Context(ctx)

	isTerminal := terminal.IsTerminal()
	terminalWidth, terminalHeight, _ := term.GetSize(0)

	installLogger(verboseFromArgs(os.Args[1:]))

	uxBlocks := uxBlock.NewBlocks(isTerminal, terminalWidth, terminalHeight, cancel)

	cliStorage, err := createCliStorage()
	if err != nil {
		printError(err, uxBlocks)
	}

	flagParams := flagParams.New()

	cobraCmd, err := buildCobraCmd(rootCmd, flagParams, uxBlocks, cliStorage)
	if err != nil {
		printError(err, uxBlocks)
	}

	cobraCmd.SetGlobalNormalizationFunc(normalizeFlagNames)

	err = cobraCmd.ExecuteContext(ctx)
	if err != nil {
		printError(err, uxBlocks)
	}
}

func printError(err error, uxBlocks uxBlock.UxBlocks) {
	if err == nil {
		return
	}
	slog.Debug("command.error", "err", fmt.Sprintf("%+v", err))

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintErrorText(err.Error())
		os.Exit(1)
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

		os.Exit(1)
	}

	if errors.Is(err, models.ErrCtrlC) {
		uxBlocks.PrintInfo(styles.InfoLine("canceled"))
		os.Exit(0)
	}

	uxBlocks.PrintErrorText(err.Error())
	os.Exit(1)
}

// installLogger assembles slog handlers (file + optional stderr) and
// installs the resulting logger as slog.Default.
func installLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}

	var handlers []slog.Handler
	if logFilePath, fileMode, err := constants.LogFilePath(); err != nil {
		// Best effort: warn directly to stderr; logger isn't up yet.
		fmt.Fprintln(os.Stderr, styles.WarningLine(err.Error()).String())
	} else {
		handlers = append(handlers, logfile.New(logfile.Config{
			Path: logFilePath,
			Mode: fileMode,
			Opts: opts,
		}))
	}
	if verbose {
		handlers = append(handlers, slog.NewTextHandler(os.Stderr, opts))
	}

	var h slog.Handler
	switch len(handlers) {
	case 0:
		h = slog.NewTextHandler(io.Discard, opts)
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
