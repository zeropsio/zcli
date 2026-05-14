package cmd

import (
	"context"
	"time"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

// LoginCmd authenticates with a Zerops token and persists it to cliStorage
// alongside the resolved region. Always runs as guest — never calls RequireAuth.
type LoginCmd struct {
	Token string `arg:"" name:"token" help:"Login token."`
	// RegionURL has no kong default — it's resolved from constants in Run.
	// Tag-level defaults need ${var} registration; the runtime fallback is
	// simpler for an internal URL no user ever overrides via flag.
	RegionURL string `name:"region-url" hidden:"" help:"Zerops region file url."`
	Region    string `name:"region" default:"prg1" hidden:"" help:"Choose one of Zerops regions. Use the \"zcli region list\" command to list all Zerops regions."`
}

func (c *LoginCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	uxBlocks := d.UxBlocks

	regionURL := c.RegionURL
	if regionURL == "" {
		// kong tag default substitution doesn't fire when the flag is bound
		// from yaml or env to an empty string — guard against it.
		regionURL = constants.DefaultRegionUrl
	}

	regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))
	regions, err := regionRetriever.RetrieveAllFromURL(ctx, regionURL)
	if err != nil {
		return err
	}

	reg, err := getLoginRegion(ctx, uxBlocks, regions, c.Region)
	if err != nil {
		return err
	}

	restApiClient := zeropsRestApiClient.NewAuthorizedClient(c.Token, reg.Address)

	response, err := restApiClient.GetUserInfo(ctx)
	if err != nil {
		return err
	}
	output, err := response.Output()
	if err != nil {
		return err
	}

	if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.Token = c.Token
		data.RegionData = reg
		return data
	}); err != nil {
		return err
	}

	uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LoginSuccess, output.FullName, output.Email)))
	return nil
}
