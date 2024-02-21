package cmdBuilder

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/params"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
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
	UxBlocks   uxBlock.UxBlocks
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

func (b *CmdBuilder) createCmdRunFunc(
	cmd *Cmd,
	params *params.Handler,
	uxBlocks uxBlock.UxBlocks,
	cliStorage *cliStorage.Handler,
) func(*cobra.Command, []string) error {
	return func(cobraCmd *cobra.Command, args []string) (err error) {
		ctx := cobraCmd.Context()

		uxBlocks.LogDebug(fmt.Sprintf("Command: %s", cobraCmd.CommandPath()))

		params.InitViper()

		argsMap, err := convertArgs(cmd, args)
		if err != nil {
			return err
		}

		guestCmdData := &GuestCmdData{
			CliStorage: cliStorage,
			UxBlocks:   uxBlocks,
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
			return nil, errors.Errorf(i18n.T(i18n.ArgsOnlyOneOptionalAllowed), arg.name)
		}
		if arg.isArray && i != len(cmd.args)-1 {
			return nil, errors.Errorf(i18n.T(i18n.ArgsOnlyOneArrayAllowed), arg.name)
		}
		if !arg.optional {
			requiredArgsCount++
		}
		isArray = arg.isArray
	}

	if len(args) < requiredArgsCount {
		return nil, errors.Errorf(i18n.T(i18n.ArgsNotEnoughRequiredArgs), requiredArgsCount, len(args))
	}

	// the last arg is not an array, max number of given args can't be greater than the number of registered args
	if !isArray && len(args) > len(cmd.args) {
		return nil, errors.Errorf(i18n.T(i18n.ArgsTooManyArgs), len(cmd.args), len(args))
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
