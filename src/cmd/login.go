package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/region"
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
		StringFlag("regionUrl", constants.DefaultRegionUrl, i18n.T(i18n.RegionUrlFlag), cmdBuilder.HiddenFlag()).
		StringFlag("region", "", i18n.T(i18n.RegionFlag), cmdBuilder.HiddenFlag()).
		HelpFlag(i18n.T(i18n.CmdHelpLogin)).
		Arg("token").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			uxBlocks := cmdData.UxBlocks

			regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))

			regions, err := regionRetriever.RetrieveAllFromURL(ctx, cmdData.Params.GetString("regionUrl"))
			if err != nil {
				return err
			}

			reg, err := getLoginRegion(ctx, uxBlocks, regions, cmdData.Params.GetString("region"))
			if err != nil {
				return err
			}

			restApiClient := zeropsRestApiClient.NewAuthorizedClient(cmdData.Args["token"][0], "https://"+reg.Address)

			response, err := restApiClient.GetUserInfo(ctx)
			if err != nil {
				return err
			}

			output, err := response.Output()
			if err != nil {
				return err
			}

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.Token = cmdData.Args["token"][0]
				data.RegionData = reg
				return data
			})
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LoginSuccess, output.FullName, output.Email)))

			return nil
		})
}

func getLoginRegion(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	regions []region.RegionItem,
	selectedRegion string,
) (region.RegionItem, error) {
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

	uxBlocks.PrintWarning(styles.WarningLine(fmt.Sprintf("Region '%s' was not found", selectedRegion)))

	header := table.NewRowFromStrings(i18n.T(i18n.RegionTableColumnName), "default")

	tableBody := table.NewBody()
	for _, reg := range regions {
		tableBody.AddStringsRow(
			reg.Name,
			fmt.Sprintf("%t", reg.IsDefault),
		)
	}

	selected, err := uxBlock.RunR(
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
		return region.RegionItem{}, err
	}

	reg := regions[selected]
	uxBlocks.PrintInfo(styles.InfoWithValueLine("Selected region", reg.Name))
	return reg, nil
}
