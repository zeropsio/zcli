package cmdBuilder

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/loginSession"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/upgrade"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type GuestCmdData struct {
	CliStorage *cliStorage.Handler
	UxBlocks   *uxBlock.Blocks
	Args       map[string][]string
	Params     flagParams.ParamsReader
	Stdout     printer.Printer
	Stderr     printer.Printer

	PrintHelp func()
}

type LoggedUserCmdData struct {
	*GuestCmdData
	RestApiClient *zeropsRestApiClient.Handler

	// optional params
	Project optional.Null[entity.Project]
	Service optional.Null[entity.Service]

	ProjectSelector func(context.Context, *LoggedUserCmdData, ...uxHelpers.ProjectSelectorOption) (entity.Project, bool, error)

	VpnKeys map[uuid.ProjectId]entity.VpnKey

	tokenSource zeropsRestApiClient.TokenSource
}

func createCmdRunFunc(
	cmd *Cmd,
	flagParams *flagParams.Handler,
	uxBlocks *uxBlock.Blocks,
	cliStorage *cliStorage.Handler,
	stdout io.Writer,
	stderr io.Writer,
) func(*cobra.Command, []string) error {
	return func(cobraCmd *cobra.Command, args []string) (err error) {
		ctx := cobraCmd.Context()

		uxBlocks.LogDebug(fmt.Sprintf("Command: %s", cobraCmd.CommandPath()))

		go upgrade.RefreshCacheIfStale(ctx)
		if warning := upgrade.NewUpgrader().MismatchWarning(); warning != "" {
			uxBlocks.PrintWarningText(warning)
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
			Stdout:     printer.NewPrinter(stdout),
			Stderr:     printer.NewPrinter(stderr),

			PrintHelp: func() {
				cobraCmd.HelpFunc()(cobraCmd, []string{})
			},
		}

		storedData := cliStorage.Data()
		if !hasToken(storedData) && cmd.guestRunFunc == nil && terminal.IsInteractive() {
			loggedIn, err := offerBrowserLogin(ctx, guestCmdData, storedData.RegionData)
			if err != nil {
				return err
			}
			if loggedIn {
				storedData = cliStorage.Data()
			}
		}

		if !hasToken(storedData) {
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
			VpnKeys:      storedData.ProjectVpnKeyRegistry,
		}

		host := storedData.RegionData.Address
		if host == "" {
			host = constants.DefaultRegion
		}
		source := tokenSource(storedData, cliStorage, host)
		cmdData.tokenSource = source
		cmdData.RestApiClient = zeropsRestApiClient.NewTokenSourceClient(source, host)

		err = func() error {
			if cmd.scopeLevel != nil {
				if err := cmd.scopeLevel.LoadSelectedScope(ctx, cmd, cmdData); err != nil {
					return err
				}
			}
			return cmd.loggedUserRunFunc(ctx, cmdData)
		}()
		// The root command reports API errors in its output and succeeds, so an expired session is checked even without err.
		if session, ok := source.(*loginSession.Source); ok && session.Expired() {
			return errorsx.NewUserError(i18n.T(i18n.LoginSessionExpired), err)
		}
		return err
	}
}

func EnvToken() string {
	return os.Getenv(constants.CliTokenEnvVar)
}

func hasToken(data cliStorage.Data) bool {
	return EnvToken() != "" || data.Token != "" || data.LoginSession.IsSet()
}

func tokenSource(data cliStorage.Data, storage *cliStorage.Handler, host string) zeropsRestApiClient.TokenSource {
	if token := EnvToken(); token != "" {
		return zeropsRestApiClient.StaticToken(token)
	}
	if data.Token != "" {
		return zeropsRestApiClient.StaticToken(data.Token)
	}
	return loginSession.New(storage, zeropsRestApiClient.NewClient(host))
}

func offerBrowserLogin(ctx context.Context, cmdData *GuestCmdData, storedRegion region.Item) (bool, error) {
	confirmed, err := uxHelpers.YesNoPrompt(ctx, i18n.T(i18n.LoginBrowserPrompt), prompt.WithCursorPosition(1))
	if err != nil || !confirmed {
		return false, err
	}
	reg, err := defaultLoginRegion(ctx, storedRegion)
	if err != nil {
		return false, err
	}
	return true, BrowserLogin(ctx, cmdData, reg)
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
