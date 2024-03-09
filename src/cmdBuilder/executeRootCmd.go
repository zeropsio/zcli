package cmdBuilder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/apiError"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

func ExecuteRootCmd(rootCmd *Cmd) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	regSignals(cancel)
	ctx = support.Context(ctx)

	isTerminal := isTerminal()

	width, _, err := term.GetSize(0)
	if err != nil {
		width = 100
	}

	outputLogger, debugFileLogger := createLoggers(isTerminal)

	uxBlocks := uxBlock.NewBlock(outputLogger, debugFileLogger, isTerminal, width, cancel)

	defer func() {
		if err != nil {
			printError(err, uxBlocks)
		}
	}()

	cliStorage, err := createCliStorage()
	if err != nil {
		return err
	}

	flagParams := flagParams.New()

	cobraCmd, err := buildCobraCmd(rootCmd, flagParams, uxBlocks, cliStorage)
	if err != nil {
		return err
	}

	err = cobraCmd.ExecuteContext(ctx)
	if err != nil {
		printError(err, uxBlocks)
	}

	return nil
}

func printError(err error, uxBlocks uxBlock.UxBlocks) {
	uxBlocks.LogDebug(fmt.Sprintf("error: %+v", err))

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintError(styles.ErrorLine(err.Error()))
		return
	}

	var apiErr apiError.Error
	if errors.As(err, &apiErr) {
		uxBlocks.PrintError(styles.ErrorLine(apiErr.GetMessage()))
		if apiErr.GetMeta() != nil {
			meta, err := yaml.Marshal(apiErr.GetMeta())
			if err != nil {
				uxBlocks.PrintError(styles.ErrorLine(fmt.Sprintf("couldn't parse meta of error: %s", apiErr.GetMessage())))
			}
			uxBlocks.PrintError(styles.ErrorLine(string(meta)))
		}

		return
	}

	uxBlocks.PrintError(styles.ErrorLine(err.Error()))
}

type terminalMode string

const (
	TerminalModeAuto     terminalMode = "auto"
	TerminalModeDisabled terminalMode = "disabled"
	TerminalModeEnabled  terminalMode = "enabled"
)

func isTerminal() bool {
	env := os.Getenv(constants.CliTerminalMode)

	switch terminalMode(env) {
	case TerminalModeAuto, "":
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	case TerminalModeDisabled:
		return false
	case TerminalModeEnabled:
		return true
	default:
		os.Stdout.WriteString(styles.WarningLine(i18n.T(i18n.UnknownTerminalMode, env)).String())

		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}
}

func createLoggers(isTerminal bool) (*logger.Handler, *logger.Handler) {
	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
	})

	loggerFilePath, err := constants.LogFilePath()
	if err != nil {
		outputLogger.Warning(styles.WarningLine(err.Error()))
	}

	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: loggerFilePath,
	})

	return outputLogger, debugFileLogger
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
	filePath, err := constants.CliDataFilePath()
	if err != nil {
		return nil, err
	}
	s, err := storage.New[cliStorage.Data](
		storage.Config{
			FilePath: filePath,
		},
	)
	return &cliStorage.Handler{Handler: s}, err
}
