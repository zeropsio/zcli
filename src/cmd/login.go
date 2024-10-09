package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/uxBlock"
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
			// Check if token is provided
			if len(cmdData.Args["token"]) == 0 {
				return errors.New("token is required")
			}
			token := cmdData.Args["token"][0]
			if token == "" {
				return errors.New("token cannot be empty")
			}

			uxBlocks := cmdData.UxBlocks

			regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))

			regionUrl := cmdData.Params.GetString("regionUrl")
			if regionUrl == "" {
				return errors.New("regionUrl is empty")
			}

			regions, err := regionRetriever.RetrieveAllFromURL(ctx, regionUrl)
			if err != nil {
				return errors.Wrap(err, "failed to retrieve regions")
			}

			if len(regions) == 0 {
				return errors.New("no regions available")
			}

			reg, err := getLoginRegion(ctx, uxBlocks, regions, cmdData.Params.GetString("region"))
			if err != nil {
				return errors.Wrap(err, "failed to get login region")
			}

			restApiClient := zeropsRestApiClient.NewAuthorizedClient(token, "https://"+reg.Address)

			response, err := restApiClient.GetUserInfo(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to get user info")
			}

			output, err := response.Output()
			if err != nil {
				return errors.Wrap(err, "failed to process user info output")
			}

			if output.FullName == "" || output.Email == "" {
				return errors.New("incomplete user info: missing full name or email")
			}

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.Token = token
				data.RegionData = reg
				return data
			})
			if err != nil {
				return errors.Wrap(err, "failed to update CLI storage")
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
	if len(regions) == 0 {
		return region.RegionItem{}, errors.New("no regions available")
	}

	if selectedRegion != "" {
		for _, reg := range regions {
			if reg.Name == selectedRegion {
				return reg, nil
			}
		}
		return region.RegionItem{}, errors.New(i18n.T(i18n.RegionNotFound, selectedRegion))
	}

	for _, reg := range regions {
		if reg.IsDefault {
			return reg, nil
		}
	}

	header := (&uxBlock.TableRow{}).AddStringCells(i18n.T(i18n.RegionTableColumnName))

	tableBody := &uxBlock.TableBody{}
	for _, reg := range regions {
		tableBody.AddStringsRow(
			reg.Name,
		)
	}

	regionIndex, err := uxBlocks.Select(
		ctx,
		tableBody,
		uxBlock.SelectLabel(i18n.T(i18n.ProjectSelectorPrompt)),
		uxBlock.SelectTableHeader(header),
	)
	if err != nil {
		return region.RegionItem{}, errors.Wrap(err, "failed to select region")
	}

	if len(regionIndex) == 0 {
		return region.RegionItem{}, errors.New("no region selected")
	}

	if regionIndex[0] < 0 || regionIndex[0] >= len(regions) {
		return region.RegionItem{}, errors.New("invalid region index selected")
	}

	return regions[regionIndex[0]], nil
}
