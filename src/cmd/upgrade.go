package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	getVersion "github.com/zeropsio/zcli/src/version"
)

func upgradeCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("upgrade").
		Short("Upgrade zcli to the latest release.").
		HelpFlag("Help for the upgrade command.").
		BoolFlag("check", false, "Print current and latest version, then exit. 0 = up to date, 1 = behind, 2 = error.").
		BoolFlag("yes", false, "Skip the confirmation prompt.").
		StringFlag("version", "", "Install a specific release tag instead of the latest.").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			check := cmdData.Params.GetBool("check")
			yes := cmdData.Params.GetBool("yes")
			targetVersion := cmdData.Params.GetString("version")

			plan, err := getVersion.PlanUpgrade(ctx, getVersion.UpgradeOptions{TargetVersion: targetVersion})
			if err != nil {
				if check {
					cmdData.Stderr.Printf("error: %s\n", err)
					os.Exit(2)
				}
				return err
			}

			if check {
				cmdData.Stdout.Printf("Current: %s\nLatest:  %s\n", plan.Current, plan.Target)
				if plan.Current == plan.Target {
					return nil
				}
				os.Exit(1)
			}

			if err := getVersion.RequireSelfUpdatable(); err != nil {
				return err
			}

			if plan.Current == plan.Target && targetVersion == "" {
				cmdData.Stdout.Printf("zcli is already on %s.\n", plan.Current)
				return nil
			}

			if !yes {
				question := fmt.Sprintf("Current: %s\nTarget:  %s\n\nUpdate?", plan.Current, plan.Target)
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
						return getVersion.Upgrade(ctx, plan)
					},
					RunningMessage:      fmt.Sprintf("Downloading and installing %s", plan.Target),
					ErrorMessageMessage: fmt.Sprintf("Upgrade to %s failed", plan.Target),
					SuccessMessage:      fmt.Sprintf("Updated to %s. Run `zcli version` to confirm.", plan.Target),
				}},
			)
		})
}
