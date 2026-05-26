package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/upgrade"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func updateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("update").
		Short(i18n.T(i18n.CmdDescUpdate)).
		HelpFlag(i18n.T(i18n.CmdHelpUpdate)).
		BoolFlag("yes", false, i18n.T(i18n.UpdateYesFlag)).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			upgrader := upgrade.NewUpgrader()
			yes := cmdData.Params.GetBool("yes")

			plan, err := upgrader.PlanUpgrade(ctx, upgrade.Options{})
			if err != nil {
				return err
			}

			if err := upgrader.RequireSelfUpdatable(); err != nil {
				return err
			}

			if err := plan.RequireSelfUpgradable(); err != nil {
				return err
			}

			if !plan.NeedsUpgrade() {
				cmdData.Stdout.Printf(
					i18n.T(i18n.UpdateAlreadyUpToDate),
					plan.Current(),
				)
				return nil
			}

			cmdData.Stdout.Printf(
				i18n.T(i18n.UpdateVersionInfo),
				plan.Current(),
				plan.Target(),
			)

			if !yes {
				confirmed, err := uxHelpers.YesNoPrompt(
					ctx,
					i18n.T(i18n.UpdatePrompt),
					prompt.WithDialogBoxStyle(styles.DialogBox()),
				)
				if err != nil {
					return err
				}
				if !confirmed {
					cmdData.Stdout.Println(i18n.T(i18n.UpdateAborted))
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
					RunningMessage:      fmt.Sprintf(i18n.T(i18n.UpdateDownloading), plan.Target()),
					ErrorMessageMessage: fmt.Sprintf(i18n.T(i18n.UpdateFailed), plan.Target()),
					SuccessMessage:      fmt.Sprintf(i18n.T(i18n.UpdateSuccess), plan.Target()),
				}},
			)
		})
}
