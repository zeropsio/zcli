package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/upgrade"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func upgradeCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("upgrade").
		Short("Upgrade zcli to the latest release.").
		HelpFlag("Help for the upgrade command.").
		BoolFlag("check", false, "Print current and latest version, then exit. 0 = up to date, 1 = behind, 2 = error, 3 = target requires install.sh.").
		BoolFlag("yes", false, "Skip the confirmation prompt.").
		StringFlag("version", "", "Install a specific release tag instead of the latest.").
		StringFlag("download-timeout", "", "Overall timeout for the binary download (Go duration, e.g. '5m', '90s'). 0 disables the timeout. Default 2m.").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			check := cmdData.Params.GetBool("check")
			yes := cmdData.Params.GetBool("yes")
			targetVersion := cmdData.Params.GetString("version")
			downloadTimeoutRaw := cmdData.Params.GetString("download-timeout")

			upgrader := upgrade.NewUpgrader()
			if downloadTimeoutRaw != "" {
				d, err := time.ParseDuration(downloadTimeoutRaw)
				if err != nil {
					return fmt.Errorf("invalid --download-timeout %q: %w", downloadTimeoutRaw, err)
				}
				upgrader = upgrader.WithDownloadTimeout(d)
			}
			plan, err := upgrader.PlanUpgrade(ctx, upgrade.Options{TargetVersion: targetVersion})
			if err != nil {
				if check {
					cmdData.Stderr.Printf("error: %s\n", err)
					return errorsx.NewExitError(2)
				}
				return err
			}

			if check {
				cmdData.Stdout.Printf("Current: %s\nLatest:  %s\n", plan.Current(), plan.Target())
				if err := plan.RequireSelfUpgradable(); err != nil {
					cmdData.Stderr.Printf("%s\n", err)
					return errorsx.NewExitError(3)
				}
				if !plan.NeedsUpgrade() {
					return nil
				}
				return errorsx.NewExitError(1)
			}

			if err := upgrader.RequireSelfUpdatable(); err != nil {
				return err
			}

			if err := plan.RequireSelfUpgradable(); err != nil {
				return err
			}

			if !plan.NeedsUpgrade() && targetVersion == "" {
				cmdData.Stdout.Printf("zcli is already on %s.\n", plan.Current())
				return nil
			}

			if !yes {
				question := fmt.Sprintf("Current: %s\nTarget:  %s\n\nUpdate?", plan.Current(), plan.Target())
				confirmed, err := uxHelpers.YesNoPrompt(
					ctx,
					question,
					prompt.WithDialogBoxStyle(styles.DialogBox()),
				)
				if err != nil {
					return err
				}
				if !confirmed {
					cmdData.Stdout.Printf("Aborted.\n")
					return nil
				}
			}

			return uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F: func(ctx context.Context, _ *uxHelpers.Process) error {
						return upgrader.Apply(ctx, plan)
					},
					RunningMessage:      fmt.Sprintf("Downloading and installing %s", plan.Target()),
					ErrorMessageMessage: fmt.Sprintf("Upgrade to %s failed", plan.Target()),
					SuccessMessage:      fmt.Sprintf("Updated to %s. Run `zcli version` to confirm.", plan.Target()),
				}},
			)
		})
}
