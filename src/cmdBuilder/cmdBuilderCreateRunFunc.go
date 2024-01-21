package cmdBuilder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/params"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/support"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/apiError"
	"gopkg.in/yaml.v3"
)

type ParamsReader interface {
	GetString(name string) string
	GetInt(name string) int
	GetBool(name string) bool
}

type CmdParamReader struct {
	cobraCmd      *cobra.Command
	paramsHandler *params.Handler
}

func newCmdParamReader(cobraCmd *cobra.Command, paramsHandler *params.Handler) *CmdParamReader {
	return &CmdParamReader{
		cobraCmd:      cobraCmd,
		paramsHandler: paramsHandler,
	}
}

func (r *CmdParamReader) GetString(name string) string {
	return r.paramsHandler.GetString(r.cobraCmd, name)
}

func (r *CmdParamReader) GetInt(name string) int {
	return r.paramsHandler.GetInt(r.cobraCmd, name)
}

func (r *CmdParamReader) GetBool(name string) bool {
	return r.paramsHandler.GetBool(r.cobraCmd, name)
}

type GuestCmdData struct {
	CliStorage *cliStorage.Handler
	UxBlocks   *uxBlock.UxBlocks
	QuietMode  QuietMode
	Args       map[string][]string
	Params     ParamsReader
}

type LoggedUserCmdData struct {
	*GuestCmdData
	RestApiClient *zeropsRestApiClient.Handler

	// optional params
	Project *entity.Project
	Service *entity.Service
}

func (b *CmdBuilder) createCmdRunFunc(cmd *Cmd, params *params.Handler) func(*cobra.Command, []string) error {
	return func(cobraCmd *cobra.Command, args []string) (err error) {
		ctx, cancel := context.WithCancel(context.Background())
		regSignals(cancel)
		ctx = support.Context(ctx)

		loggerFilePath, err := constants.LogFilePath()
		if err != nil {
			return errors.New(i18n.T(i18n.LoggerUnableToOpenLogFileWarning))
		}

		isTerminal, err := isTerminal()
		if err != nil {
			return err
		}

		outputLogger, debugFileLogger := createLoggers(isTerminal, loggerFilePath)

		uxBlocks := uxBlock.NewBlock(outputLogger, debugFileLogger, isTerminal, cancel)

		uxBlocks.PrintDebugLine(fmt.Sprintf("Command: %s", cobraCmd.CommandPath()))

		defer func() {
			if err != nil {
				printError(err, uxBlocks)
				err = skipErr
			}
		}()

		err = params.InitViper()
		if err != nil {
			return err
		}

		quietMode, err := getQuietMode(isTerminal)
		if err != nil {
			return err
		}

		cliStorage, err := createCliStorage()
		if err != nil {
			return err
		}

		argsMap, err := convertArgs(cmd, args)
		if err != nil {
			return err
		}

		guestCmdData := &GuestCmdData{
			CliStorage: cliStorage,
			UxBlocks:   uxBlocks,
			QuietMode:  quietMode,
			Args:       argsMap,
			Params:     newCmdParamReader(cobraCmd, params),
		}

		if cmd.loggedUserRunFunc != nil {
			storedData := cliStorage.Data()

			token := storedData.Token
			if token == "" {
				return errors.New(i18n.T(i18n.UnauthenticatedUser))
			}

			cmdData := &LoggedUserCmdData{

				GuestCmdData: guestCmdData,
			}

			cmdData.RestApiClient = zeropsRestApiClient.NewAuthorizedClient(token, storedData.RegionData.RestApiAddress)

			for _, dep := range getDependencyListFromRoot(cmd.scopeLevel) {
				err := dep.LoadSelectedScope(ctx, cmd, cmdData)
				if err != nil {
					return err
				}
			}
			return cmd.loggedUserRunFunc(ctx, cmdData)

		}

		return cmd.guestRunFunc(ctx, guestCmdData)
	}
}

func convertArgs(cmd *Cmd, args []string) (map[string][]string, error) {
	var requiredArgsCount int
	var isArray bool
	for i, arg := range cmd.args {
		if arg.optional && i != len(cmd.args)-1 {
			return nil, errors.Errorf("optional arg %s can be only the last one", arg.name)
		}
		if arg.isArray && i != len(cmd.args)-1 {
			return nil, errors.Errorf("array arg %s can be only the last one", arg.name)
		}
		if !arg.optional {
			requiredArgsCount++
		}
		isArray = arg.isArray
	}

	if len(args) < requiredArgsCount {
		// TODO - janhajek message
		return nil, errors.Errorf("expected at least %d arg(s), got %d", requiredArgsCount, len(args))
	}

	// the last arg is not an array, max number of given args can't be greater than the number of registered args
	if !isArray && len(args) > len(cmd.args) {
		// TODO - janhajek message
		return nil, errors.Errorf("expected no more than %d arg(s), got %d", len(cmd.args), len(args))
	}

	argsMap := make(map[string][]string)
	for i, arg := range cmd.args {
		if len(args) > i {
			if arg.isArray {
				argsMap[arg.name] = args[i:]
			} else {
				argsMap[arg.name] = []string{args[i]}
			}
		}
	}

	return argsMap, nil
}

func printError(err error, uxBlocks *uxBlock.UxBlocks) {
	uxBlocks.PrintDebugLine(fmt.Sprintf("error: %+v", err))

	if userErr := errorsx.AsUserError(err); userErr != nil {
		uxBlocks.PrintErrorLine(err.Error())
		return
	}

	var apiErr apiError.Error
	if errors.As(err, &apiErr) {
		uxBlocks.PrintErrorLine(apiErr.GetMessage())
		if apiErr.GetMeta() != nil {
			meta, err := yaml.Marshal(apiErr.GetMeta())
			if err != nil {
				uxBlocks.PrintErrorLine(fmt.Sprintf("couldn't parse meta of error: %s", apiErr.GetMessage()))
			}
			uxBlocks.PrintErrorLine(string(meta))
		}

		return
	}

	uxBlocks.PrintErrorLine(err.Error())
}

func getQuietMode(isTerminal bool) (QuietMode, error) {
	if !isTerminal {
		return QuietModeConfirmNothing, nil
	}

	switch QuietMode(QuietModeFlag) {
	case QuietModeConfirmNothing, QuietModeConfirmAll, QuietModeConfirmOnlyDestructive:
		return QuietMode(QuietModeFlag), nil
	default:
		// TODO - janhajek message
		return 0, errors.New("unknown quiet mode")
	}
}

func isTerminal() (bool, error) {
	switch TerminalMode(TerminalFlag) {
	case TerminalModeAuto:
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()), nil
	case TerminalModeDisabled:
		return false, nil
	case TerminalModeEnabled:
		return true, nil
	default:
		// TODO - janhajek message
		return false, errors.New("unknown terminal mode")
	}
}

func createLoggers(isTerminal bool, logFilePathFlag string) (*logger.Handler, *logger.Handler) {
	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
	})

	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: logFilePathFlag,
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
