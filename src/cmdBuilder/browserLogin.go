package cmdBuilder

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/browser"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/deviceLogin"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/loginSession"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

// BrowserLogin runs the device login against reg and stores the session in place of any stored token.
func BrowserLogin(ctx context.Context, cmdData *GuestCmdData, reg region.Item) error {
	interactive := terminal.IsInteractive()
	if !interactive && isCI() {
		return errorsx.NewUserError(i18n.T(i18n.LoginBrowserInCi, constants.CliTokenEnvVar), nil)
	}

	auth, err := deviceLogin.Run(ctx, zeropsRestApiClient.NewClient(reg.Address), deviceLoginConfig(cmdData.UxBlocks, interactive))
	if err != nil {
		return err
	}
	session, err := loginSession.FromAuth(auth.Auth)
	if err != nil {
		return err
	}

	_, err = cmdData.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
		data.Token = ""
		data.LoginSession = session
		data.RegionData = reg
		return data
	})
	if err != nil {
		return err
	}

	fullName, email := "", ""
	if auth.User != nil {
		fullName, email = auth.User.FullName.Native(), auth.User.Email.Native()
	}
	cmdData.UxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.LoginSuccess, fullName, email)))
	return nil
}

func deviceLoginConfig(uxBlocks uxBlock.UxBlocks, interactive bool) deviceLogin.Config {
	config := deviceLogin.Config{
		AppCode:     constants.DeviceLoginAppCode,
		Interactive: interactive,
		UxBlocks:    uxBlocks,
	}
	// Scripts and agents without a terminal pass the printed URL on, a browser window there would be unexpected.
	if interactive {
		config.OpenBrowser = browser.Open
	}
	return config
}

// isCI follows the CI env variable convention of GitHub Actions, GitLab CI, CircleCI, Bitbucket, Buildkite and Travis.
func isCI() bool {
	value := os.Getenv("CI")
	if value == "" {
		return false
	}
	isSet, err := strconv.ParseBool(value)
	return err != nil || isSet
}

// defaultLoginRegion is the region for a login that a command other than `zcli login` starts.
func defaultLoginRegion(ctx context.Context, stored region.Item) (region.Item, error) {
	if stored.Address != "" {
		return stored, nil
	}
	regions, err := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute})).RetrieveAllFromURL(ctx, constants.DefaultRegionUrl)
	if err != nil {
		return region.Item{}, err
	}
	for _, reg := range regions {
		if reg.IsDefault {
			return reg, nil
		}
	}
	return region.Item{}, errors.New("no default region found")
}
