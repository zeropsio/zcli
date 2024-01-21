package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/errorCode"
)

func loginCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("login").
		Short(i18n.T(i18n.CmdLogin)).
		StringFlag("regionUrl", constants.DefaultRegionUrl, i18n.T(i18n.RegionUrlFlag), cmdBuilder.HiddenFlag()).
		StringFlag("region", "", i18n.T(i18n.RegionFlag), cmdBuilder.HiddenFlag()).
		Arg("token").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			uxBlocks := cmdData.UxBlocks

			regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))

			regions, err := regionRetriever.RetrieveAllFromURL(cmdData.Params.GetString("regionUrl"))
			if err != nil {
				return err
			}

			reg, err := getLoginRegion(ctx, uxBlocks, regions, cmdData.Params.GetString("region"))
			if err != nil {
				return err
			}

			restApiClient := zeropsRestApiClient.NewAuthorizedClient(cmdData.Args["token"][0], reg.RestApiAddress)

			response, err := restApiClient.GetUserInfo(ctx)
			if err != nil {
				return err
			}

			output, err := response.Output()
			if err != nil {
				return zeropsRestApiClient.CheckError(
					err,
					zeropsRestApiClient.CheckErrorCode(
						errorCode.NotAuthorized,
						errorsx.NewUserError(i18n.T(i18n.LoginIncorrectToken), err),
					),
				)
			}

			_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.Token = cmdData.Args["token"][0]
				data.RegionData = reg
				return data
			})
			if err != nil {
				return err
			}

			uxBlocks.PrintSuccessLine(i18n.T(i18n.LoginSuccess, output.FullName, output.Email))

			return nil
		})
}

func getLoginRegion(
	ctx context.Context,
	uxBlocks *uxBlock.UxBlocks,
	regions []region.Data,
	selectedRegion string,
) (region.Data, error) {
	if selectedRegion != "" {
		for _, reg := range regions {
			if reg.Name == selectedRegion {
				return reg, nil
			}
		}
		return region.Data{}, errors.New(i18n.T(i18n.RegionNotFound, selectedRegion))
	}

	for _, reg := range regions {
		if reg.IsDefault {
			return reg, nil
		}
	}

	// TODO - janhajek translation
	header := (&uxBlock.TableRow{}).AddStringCells("Name")

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
		return region.Data{}, err
	}

	return regions[regionIndex[0]], nil
}
