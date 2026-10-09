package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

func loginCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("login").
		Short(i18n.T(i18n.CmdDescLogin)).
		StringFlag("region-url", constants.DefaultRegionUrl, i18n.T(i18n.RegionUrlFlag), cmdBuilder.HiddenFlag()).
		StringFlag("region", "prg1", i18n.T(i18n.RegionFlag), cmdBuilder.HiddenFlag()).
		HelpFlag(i18n.T(i18n.CmdHelpLogin)).
		Arg("token", cmdBuilder.OptionalArg()).
		BoolFlag("browser", false, i18n.T(i18n.LoginBrowserFlag)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			uxBlocks := cmdData.UxBlocks

			tokenArg, hasTokenArg := cmdData.Args["token"]
			useBrowser := cmdData.Params.GetBool("browser")
			if hasTokenArg && useBrowser {
				return errorsx.NewUserError(i18n.T(i18n.LoginTokenAndBrowser), nil)
			}
			if hasTokenArg && strings.TrimSpace(tokenArg[0]) == "" {
				return errorsx.NewUserError(i18n.T(i18n.LoginEmptyToken), nil)
			}

			regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))

			regions, err := regionRetriever.RetrieveAllFromURL(ctx, cmdData.Params.GetString("region-url"))
			if err != nil {
				return err
			}

			reg, err := getLoginRegion(ctx, uxBlocks, regions, cmdData.Params.GetString("region"))
			if err != nil {
				return err
			}

			envToken := cmdBuilder.EnvToken()
			switch {
			case hasTokenArg:
				return tokenLogin(ctx, cmdData, reg, tokenArg[0], true)
			case envToken != "" && !useBrowser:
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.LoginEnvToken, constants.CliTokenEnvVar)))
				return tokenLogin(ctx, cmdData, reg, envToken, false)
			case envToken != "":
				uxBlocks.PrintWarning(styles.WarningLine(i18n.T(i18n.LoginEnvOverrides, constants.CliTokenEnvVar)))
			}
			return cmdBuilder.BrowserLogin(ctx, cmdData, reg)
		})
}

// tokenLogin verifies the token and stores the region, plus the token itself when persistToken is set.
func tokenLogin(ctx context.Context, cmdData *cmdBuilder.GuestCmdData, reg region.Item, token string, persistToken bool) error {
	response, err := zeropsRestApiClient.NewAuthorizedClient(token, reg.Address).GetUserInfo(ctx)
	if err != nil {
		return err
	}
	output, err := response.Output()
	if err != nil {
		return err
	}

	_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		if persistToken {
			data.Token = token
			data.LoginSession = cliStorage.LoginSession{}
		}
		data.RegionData = reg
		return data
	})
	if err != nil {
		return err
	}

	cmdData.UxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LoginSuccess, output.FullName, output.Email)))
	return nil
}

func getLoginRegion(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	regions []region.Item,
	selectedRegion string,
) (region.Item, error) {
	if selectedRegion == "" {
		for _, reg := range regions {
			if reg.IsDefault {
				return reg, nil
			}
		}
	}

	if selectedRegion != "" {
		for _, reg := range regions {
			if reg.Name == selectedRegion {
				return reg, nil
			}
		}
	}

	regionNotFoundErr := errors.Errorf("Region '%s' was not found", selectedRegion)
	if !terminal.IsTerminal() {
		return region.Item{}, regionNotFoundErr
	}

	uxBlocks.PrintWarning(styles.WarningLine(regionNotFoundErr.Error()))

	header := table.NewRowFromStrings("name", "default")

	tableBody := table.NewBody()
	for _, reg := range regions {
		tableBody.AddStringsRow(
			reg.Name,
			fmt.Sprintf("%t", reg.IsDefault),
		)
	}

	selected, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			tableBody,
			selector.WithLabel("Select region"),
			selector.WithHeader(header),
			selector.WithEnableFiltering(),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return region.Item{}, err
	}

	reg := regions[selected]
	uxBlocks.PrintInfo(styles.InfoWithValueLine("Selected region", reg.Name))
	return reg, nil
}
