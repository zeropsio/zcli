// Helpers extracted from the cobra-side command files. These survive the
// cobra → kong cutover because the kong commands still depend on them.
//
// Sourced from (pre-deletion):
//   - disconnectVpn, isVpnUp, vpnCheckAddress: vpnDown.go / vpnUp.go
//   - getLoginRegion: login.go
//   - guestInfoPart: env.go
//   - Env (struct): projectEnv.go
//   - maxEnvFileSize (const): serviceCreate.go
//
// Will fold back into the migrated command files (or get dropped) when the
// remaining cobra cmdBuilder framework retires — which is gated on
// `project scope` migrating.

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmdRunner"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/nettools"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/units"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/wg"
)

const vpnCheckAddress = "logger.core.zerops"

const maxEnvFileSize = units.MiB

// Env is the per-line template input for `project env`.
type Env struct {
	Key, Value string
}

func disconnectVpn(ctx context.Context, uxBlocks uxBlock.UxBlocks, dnsSetup, checkInstallation bool) error {
	if err := wg.CheckWgInstallation(checkInstallation, dnsSetup); err != nil {
		return err
	}
	filePath, fileMode, err := constants.WgConfigFilePath()
	if err != nil {
		return err
	}
	// wg-quick needs the file to exist; empty is fine.
	f, err := file.Open(filePath, os.O_RDWR|os.O_CREATE, fileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := cmdRunner.Run(wg.DownCmd(ctx, filePath, constants.WgInterfaceName)); err != nil {
		return err
	}
	uxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.VpnDown)))
	return nil
}

func isVpnUp(ctx context.Context, uxBlocks *uxBlock.Blocks, attempts int) bool {
	err := uxHelpers.ProcessCheckWithSpinner(ctx, uxBlocks, []uxHelpers.Process{{
		F: func(ctx context.Context, _ *uxHelpers.Process) error {
			for i := 0; i < attempts; i++ {
				if err := nettools.Ping(ctx, vpnCheckAddress); err == nil {
					return nil
				}
				time.Sleep(500 * time.Millisecond)
			}
			return errors.New(i18n.T(i18n.VpnPingFailed))
		},
		RunningMessage: i18n.T(i18n.VpnCheckingConnection),
	}})
	return err == nil
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
		tableBody.AddStringsRow(reg.Name, fmt.Sprintf("%t", reg.IsDefault))
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

func guestInfoPart(tableBody *table.Body) {
	cliDataFilePath, _, err := constants.CliDataFilePath()
	if err != nil {
		cliDataFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoCliDataFilePath), cliDataFilePath)

	logFilePath, _, err := constants.LogFilePath()
	if err != nil {
		logFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoLogFilePath), logFilePath)

	wgConfigFilePath, _, err := constants.WgConfigFilePath()
	if err != nil {
		wgConfigFilePath = err.Error()
	}
	tableBody.AddStringsRow(i18n.T(i18n.StatusInfoWgConfigFilePath), wgConfigFilePath)
}
