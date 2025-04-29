package cmdBuilder

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock"
	getVersion "github.com/zeropsio/zcli/src/version"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type ParamsReader interface {
	GetString(name string) string
	GetInt(name string) int
	GetBool(name string) bool
	IsSet(name string) bool
}

type GuestCmdData struct {
	CliStorage *cliStorage.Handler
	UxBlocks   *uxBlock.Blocks
	Args       map[string][]string
	Params     ParamsReader
	Stdout     printer.Printer
	Stderr     printer.Printer

	PrintHelp func()
}

type LoggedUserCmdData struct {
	*GuestCmdData
	RestApiClient *zeropsRestApiClient.Handler

	// optional params
	Project *entity.Project
	Service *entity.Service

	VpnKeys map[uuid.ProjectId]entity.VpnKey
}

func createCmdRunFunc(
	cmd *Cmd,
	flagParams *flagParams.Handler,
	uxBlocks *uxBlock.Blocks,
	cliStorage *cliStorage.Handler,
) func(*cobra.Command, []string) error {
	return func(cobraCmd *cobra.Command, args []string) (err error) {
		ctx := cobraCmd.Context()

		uxBlocks.LogDebug(fmt.Sprintf("Command: %s", cobraCmd.CommandPath()))

		if getVersion.IsVersionCheckMismatch(ctx) {
			versionCheckMismatch, err := getVersion.GetVersionCheckMismatch()
			if err != nil {
				return err
			}
			uxBlocks.PrintWarningText(versionCheckMismatch)
		}

		flagParams.Bind(cobraCmd)

		argsMap, err := convertArgs(cmd, args)
		if err != nil {
			return err
		}

		guestCmdData := &GuestCmdData{
			CliStorage: cliStorage,
			UxBlocks:   uxBlocks,
			Args:       argsMap,
			Params:     flagParams,
			Stdout:     printer.NewPrinter(os.Stdout),
			Stderr:     printer.NewPrinter(os.Stderr),

			PrintHelp: func() {
				cobraCmd.HelpFunc()(cobraCmd, []string{})
			},
		}

		storedData := cliStorage.Data()

		token := storedData.Token
		if token == "" {
			if cmd.guestRunFunc != nil {
				return cmd.guestRunFunc(ctx, guestCmdData)
			}
			return errors.New(i18n.T(i18n.UnauthenticatedUser))
		}

		// user is logged in but there is only the guest run func
		if cmd.loggedUserRunFunc == nil {
			return cmd.guestRunFunc(ctx, guestCmdData)
		}

		cmdData := &LoggedUserCmdData{
			GuestCmdData: guestCmdData,
			VpnKeys:      storedData.VpnKeys,
		}

		cmdData.RestApiClient = zeropsRestApiClient.NewAuthorizedClient(token, "https://"+storedData.RegionData.Address)

		if cmd.scopeLevel != nil {
			if err := cmd.scopeLevel.LoadSelectedScope(ctx, cmd, cmdData); err != nil {
				return err
			}
		}

		return cmd.loggedUserRunFunc(ctx, cmdData)
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
