package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/upgrade"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
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
		BoolFlag("no-cache", false, "Bypass the on-disk version cache and resolve `latest` directly from the release API.").
		BoolFlag("pick-version", false, "Open an interactive picker listing every release. Pre-v1.1.0 entries are shown but disabled (use install.sh for those).").
		BoolFlag("include-pre-release", false, "Include pre-release/rc tags in the --pick-version picker (default: stable releases only).").
		StringFlag("version", "", "Install a specific release tag instead of the latest.").
		StringFlag("download-timeout", "", "Overall timeout for the binary download (Go duration, e.g. '5m', '90s'). 0 disables the timeout. Default 2m.").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			check := cmdData.Params.GetBool("check")
			yes := cmdData.Params.GetBool("yes")
			noCache := cmdData.Params.GetBool("no-cache")
			pickVersion := cmdData.Params.GetBool("pick-version")
			includePrerelease := cmdData.Params.GetBool("include-pre-release")
			targetVersion := cmdData.Params.GetString("version")
			downloadTimeoutRaw := cmdData.Params.GetString("download-timeout")

			if pickVersion && targetVersion != "" {
				return errors.New("--pick-version and --version are mutually exclusive")
			}

			upgrader := upgrade.NewUpgrader()
			// --download-timeout uses StringFlag because cmdBuilder has no DurationFlag; parse manually and leave Upgrader's default in place when unset.
			if downloadTimeoutRaw != "" {
				d, err := time.ParseDuration(downloadTimeoutRaw)
				if err != nil {
					return fmt.Errorf("invalid --download-timeout %q: %w", downloadTimeoutRaw, err)
				}
				upgrader = upgrader.WithDownloadTimeout(d)
			}
			if pickVersion {
				picked, err := pickReleaseInteractive(ctx, upgrader, includePrerelease)
				if err != nil {
					return err
				}
				// Pre-v1.1.0 picks can't be applied by `zcli upgrade`; show the
				// install.sh fallback as a warning and stop. The picker keeps these
				// rows enabled (vs. disabled) so the user gets here intentionally
				// and the tag is already in their hands when they want to paste it
				// into install.sh.
				if !picked.SelfUpgradable {
					cmdData.UxBlocks.PrintWarningText(upgrade.InstallScriptHint(picked.Tag))
					return nil
				}
				targetVersion = picked.Tag
			}
			plan, err := upgrader.PlanUpgrade(ctx, upgrade.Options{
				TargetVersion: targetVersion,
				NoCache:       noCache,
			})
			if err != nil {
				// --check is a scripting interface, so its errors translate to a fixed exit code instead of bubbling up to the styled error printer.
				if check {
					cmdData.Stderr.Printf("error: %s\n", err)
					return errorsx.NewExitError(2)
				}
				return err
			}

			if check {
				cmdData.Stdout.Printf("Current: %s\n", plan.Current())
				if targetVersion != "" {
					// PlanUpgrade only fetched the user-supplied target; look up the actual latest separately so the user sees all three lines.
					cmdData.Stdout.Printf("Target:  %s\n", plan.Target())
					latest, _ := upgrader.LatestTag(ctx, noCache)
					if latest != "" {
						cmdData.Stdout.Printf("Latest:  %s\n", latest)
					}
				} else {
					cmdData.Stdout.Printf("Latest:  %s\n", plan.Target())
				}
				if err := plan.RequireSelfUpgradable(); err != nil {
					cmdData.UxBlocks.PrintWarningText(err.Error())
					return errorsx.NewExitError(3)
				}
				if !plan.NeedsUpgrade() {
					return nil
				}
				return errorsx.NewExitError(1)
			}

			// Channel gate before target gate: a package-managed install can't be helped by either, and pointing at the package manager is more actionable
			// than pointing at install.sh.
			if err := upgrader.RequireSelfUpdatable(); err != nil {
				return err
			}

			if err := plan.RequireSelfUpgradable(); err != nil {
				return err
			}

			// Both conditions matter: without --version, NeedsUpgrade==false means "nothing to do". With --version the user pinned a specific tag (often a
			// downgrade), so proceed even when NeedsUpgrade returns false.
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

// pickReleaseInteractive fetches the release list and runs a selector TUI
// over it. Every entry is selectable, including pre-v1.1.0 ones: those
// rows still get a "requires install.sh" label so users see what they're
// picking, and the caller handles the post-pick branch (print the
// install.sh hint instead of going through Apply). Page size is capped at
// 15 rows so long histories paginate predictably instead of swallowing
// the whole terminal.
func pickReleaseInteractive(ctx context.Context, upgrader upgrade.Upgrader, includePrerelease bool) (upgrade.Release, error) {
	releases, err := upgrader.AvailableReleases(ctx, includePrerelease)
	if err != nil {
		return upgrade.Release{}, err
	}
	if len(releases) == 0 {
		return upgrade.Release{}, errors.New("no releases available")
	}

	header := table.NewRowFromStrings("Version", "Status")
	body := table.NewBody()
	for _, r := range releases {
		status := "stable"
		switch {
		case !r.SelfUpgradable:
			status = "requires install.sh"
		case r.Prerelease:
			status = "pre-release"
		}
		body.AddRow(table.NewRowFromStrings(r.Tag, status))
	}

	idx, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel("Pick a release to install"),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
			selector.WithMaxRowsPerPage(15),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return upgrade.Release{}, err
	}
	if idx < 0 || idx >= len(releases) {
		return upgrade.Release{}, errors.New("invalid release selection")
	}
	return releases[idx], nil
}
