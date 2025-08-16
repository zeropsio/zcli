package cmdBuilder

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/flagParams"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxHelpers"
	getVersion "github.com/zeropsio/zcli/src/version"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

const (
	EnvTokenKey = "ZEROPS_TOKEN" //nolint:gosec // Environment variable name, not a credential
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
}

// resolveRegion finds a region by name or returns the default region.
// This is a simplified, non-interactive version of the login region logic.
func resolveRegion(regions []region.Item, selectedRegion string) (region.Item, error) {
	if selectedRegion == "" {
		for _, reg := range regions {
			if reg.IsDefault {
				return reg, nil
			}
		}
		return region.Item{}, errors.New("no default region available")
	}

	for _, reg := range regions {
		if reg.Name == selectedRegion {
			return reg, nil
		}
	}

	return region.Item{}, errors.Errorf("region '%s' not found", selectedRegion)
}

func resolveAuthenticationData(ctx context.Context, storedData cliStorage.Data, storage *cliStorage.Handler) (string, region.Item, error) {
	token := storedData.Token
	regionData := storedData.RegionData

	envToken := os.Getenv(EnvTokenKey)
	usingEnvToken := false

	if token == "" && envToken != "" {
		token = envToken
		usingEnvToken = true
	}

	if token == "" {
		return "", region.Item{}, errors.New(i18n.T(i18n.UnauthenticatedUser))
	}

	// Use default region for env tokens OR when region data is missing
	if usingEnvToken || regionData.Address == "" {
		// Use the same region fetching logic as login command
		regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))
		regions, err := regionRetriever.RetrieveAllFromURL(ctx, constants.DefaultRegionUrl)
		if err != nil {
			return "", region.Item{}, errors.Wrap(err, "failed to retrieve regions")
		}

		// Use our simplified region resolution (no interactive mode)
		resolvedRegion, err := resolveRegion(regions, "")
		if err != nil {
			if usingEnvToken {
				return "", region.Item{}, errors.Wrap(err, "environment token requires default region")
			}
			return "", region.Item{}, errors.Wrap(err, "failed to resolve region data")
		}
		regionData = resolvedRegion

		// Auto-login: Save env token and region like login command does
		if usingEnvToken {
			_, err = storage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.Token = token
				data.RegionData = regionData
				return data
			})
			if err != nil {
				return "", region.Item{}, errors.Wrap(err, "failed to save authentication data")
			}
		}
	}

	return token, regionData, nil
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

		token, regionData, err := resolveAuthenticationData(ctx, storedData, cliStorage)
		if err != nil {
			if cmd.guestRunFunc != nil {
				return cmd.guestRunFunc(ctx, guestCmdData)
			}
			return err
		}

		// user is logged in but there is only the guest run func
		if cmd.loggedUserRunFunc == nil {
			return cmd.guestRunFunc(ctx, guestCmdData)
		}

		cmdData := &LoggedUserCmdData{
			GuestCmdData: guestCmdData,
			VpnKeys:      storedData.VpnKeys,
		}

		cmdData.RestApiClient = zeropsRestApiClient.NewAuthorizedClient(token, "https://"+regionData.Address)

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
