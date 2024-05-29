package cmdBuilder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/terminal"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/apiError"
)

func ExecuteRootCmd(rootCmd *Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	regSignals(cancel)
	ctx = support.Context(ctx)

	isTerminal := terminal.IsTerminal()

	width, _, err := term.GetSize(0)
	if err != nil {
		width = 100
	}

	outputLogger, debugFileLogger := createLoggers(isTerminal)

	uxBlocks := uxBlock.NewBlock(outputLogger, debugFileLogger, isTerminal, width, cancel)

	cliStorage, err := createCliStorage()
	if err != nil {
		printError(err, uxBlocks)
	}

	flagParams := flagParams.New()

	cobraCmd, err := buildCobraCmd(rootCmd, flagParams, uxBlocks, cliStorage)
	if err != nil {
		printError(err, uxBlocks)
	}

	err = cobraCmd.ExecuteContext(ctx)
	if err != nil {
		printError(err, uxBlocks)
	}
}

func printError(err error, uxBlocks uxBlock.UxBlocks) {
	if err == nil {
		return
	}
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
	os.Exit(1)
}

func createLoggers(isTerminal bool) (*logger.Handler, *logger.Handler) {
	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
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
