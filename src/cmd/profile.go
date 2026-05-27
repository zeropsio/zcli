package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func profileCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("profile").
		Short("Manage per-repository CLI profiles.").
		HelpFlag("Help for the 'profile' command.").
		AddChildrenCmd(profileInitCmd()).
		AddChildrenCmd(profileWhereCmd()).
		AddChildrenCmd(profileListCmd()).
		AddChildrenCmd(profileAddCmd()).
		AddChildrenCmd(profileCreateCmd()).
		AddChildrenCmd(profileRemoveCmd()).
		AddChildrenCmd(profileRenameCmd()).
		AddChildrenCmd(profileUseCmd())
}

func profileInitCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("init").
		Short("Create a per-repo profile at <repo>/.zcli/profile.json.").
		HelpFlag("Help for the 'profile init' command.").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			repoRoot, err := cmdData.Profile.InitRepo()
			if err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Created repo profile", profile.RepoProfilePath(repoRoot)))

			added, err := ensureGitignoreEntry(repoRoot, profile.RepoProfileDir+"/")
			if err != nil {
				cmdData.UxBlocks.PrintWarningTextf("could not update %s/.gitignore (%s); add `.zcli/` manually to keep this profile out of source control", repoRoot, err)
				return nil
			}
			if added {
				cmdData.UxBlocks.PrintInfo(styles.InfoLine("added `.zcli/` to .gitignore"))
			}
			return nil
		})
}

func profileListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short("List repo profiles, marking the active one with *.").
		HelpFlag("Help for the 'profile list' command.").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			profiles := cmdData.Profile.RepoProfiles()
			if len(profiles) == 0 {
				return errors.New("no repo profile in this directory tree; run `zcli profile init` here first")
			}
			active := cmdData.Profile.ActiveRepoProfileName()
			for _, p := range profiles {
				marker := "  "
				if p.Name == active {
					marker = "* "
				}
				cmdData.Stdout.Println(marker + p.Name)
			}
			return nil
		})
}

func profileAddCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("add").
		Short("Add a new named profile to the current repo (with a fresh id).").
		HelpFlag("Help for the 'profile add' command.").
		Arg("name").
		BoolFlag("use", false, "Switch the active repo profile to the newly created one in the same step.").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			name := cmdData.Args["name"][0]
			created, err := cmdData.Profile.AddRepoProfile(name)
			if err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Added profile", created.Name+"  (id "+created.Id+")"))
			if cmdData.Params.GetBool("use") {
				if err := cmdData.Profile.UseRepoProfile(created.Name); err != nil {
					return err
				}
				cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Active profile", created.Name))
				return nil
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoLine("active profile unchanged; run `zcli profile use " + created.Name + "` to switch (or re-run with --use)"))
			return nil
		})
}

// profileSource is the first-step choice in the interactive `profile create` builder. Picking
// a source decides whether the new entry copies its non-sensitive overrides from somewhere or
// starts clean. Tokens / VPN keys are never copied; secrets stay isolated per-profile by
// design.
type profileSource int

const (
	profileSourceScratch profileSource = iota
	profileSourceGlobal
	profileSourceExisting
)

func profileCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short("Interactively build a new named profile in the current repo.").
		HelpFlag("Help for the 'profile create' command.").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			if !cmdData.Profile.HasRepo() {
				return errors.New("`zcli profile create` only works inside a repo with a profile; run `zcli profile init` here first")
			}
			if !terminal.IsTerminal() {
				return errors.New("`zcli profile create` is interactive; run it in a real terminal, or use `zcli profile add <name>` for the one-shot flow")
			}

			source, sourceName, err := promptProfileSource(ctx, cmdData)
			if err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected source", profileSourceLabel(source, sourceName)))

			name, err := promptProfileName(ctx, cmdData.Profile.RepoProfiles())
			if err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected name", name))

			regionOverride, err := promptProfileRegion(ctx)
			if err != nil {
				return err
			}
			if regionOverride != nil {
				cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Region override", regionOverride.Name))
			} else {
				cmdData.UxBlocks.PrintInfo(styles.InfoLine("Region: inherited from global"))
			}

			cmdData.UxBlocks.PrintInfoText("Remaining details (scope, token) will be wired up in follow-up steps.")
			return nil
		})
}

// promptProfileRegion is the third step of the builder. The user picks whether the new entry
// inherits global at read time (no override) or pins a specific region listed by the regions
// endpoint. Returns nil when no override should be set.
func promptProfileRegion(ctx context.Context) (*profile.Region, error) {
	choices := []string{"Inherit global region", "Pick a different region"}
	descriptions := []string{
		"Leave the new entry's region field empty. Manager.Region() falls through to the global profile at read time, so the new profile follows whatever the user has globally - including any later `zcli login` that flips global.",
		"Pin a specific region as a repo-profile override. Reads from this profile use the chosen region regardless of what global is, which is useful when one repo is permanently against a non-default region.",
	}
	idx, err := uxBlock.Run(
		prompt.NewRoot(ctx, "How should this profile resolve its region?", choices, prompt.WithChoiceDescriptions(descriptions)),
		prompt.GetChoiceCursor,
	)
	if err != nil {
		return nil, err
	}
	if idx == 0 {
		return nil, nil
	}

	regionRetriever := region.New(httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Minute * 5}))
	regions, err := regionRetriever.RetrieveAllFromURL(ctx, constants.DefaultRegionUrl)
	if err != nil {
		return nil, errors.Wrap(err, "fetch available regions")
	}
	if len(regions) == 0 {
		return nil, errors.New("regions endpoint returned no entries; can't pick an override")
	}

	header := table.NewRowFromStrings("name", "default")
	body := table.NewBody()
	for _, r := range regions {
		body.AddStringsRow(r.Name, strconv.FormatBool(r.IsDefault))
	}
	selected, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel("Select region for this profile"),
			selector.WithHeader(header),
			selector.WithEnableFiltering(),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return nil, err
	}
	if selected < 0 || selected >= len(regions) {
		return nil, errors.New("invalid region selection")
	}
	pinned := profile.RegionFromItem(regions[selected])
	return &pinned, nil
}

// promptProfileName asks for the new profile's name with the same input model the rest of
// the CLI uses for create flows. Re-prompts on empty input and on a collision with any entry
// already in the repo file - both states are recoverable, so we stay inside the model instead
// of bailing the whole command.
func promptProfileName(ctx context.Context, existing []profile.RepoProfile) (string, error) {
	taken := make(map[string]struct{}, len(existing))
	for _, p := range existing {
		taken[p.Name] = struct{}{}
	}

	label := styles.NewStringBuilder()
	label.WriteString("Type ")
	label.WriteStyledString(styles.SelectStyle().Bold(true), "profile")
	label.WriteString(" name")

	for {
		name, err := uxBlock.Run(
			input.NewRoot(
				ctx,
				input.WithLabel(label.String()),
				input.WithHelpPlaceholder(),
				input.WithPlaceholderStyle(styles.HelpStyle()),
				input.WithoutPrompt(),
			),
			input.GetValueFunc,
		)
		if err != nil {
			return "", err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			// Stay in the loop. Print a hint so the user sees why the prompt repeated.
			continue
		}
		if _, dup := taken[name]; dup {
			return "", errors.New("profile " + strconv.Quote(name) + " already exists in this repo; pick a different name")
		}
		return name, nil
	}
}

// promptProfileSource walks the user through the source-selection step of the builder. Returns
// the chosen source and, when "from existing" is selected, the source profile's name (empty
// otherwise).
func promptProfileSource(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) (profileSource, string, error) {
	choices := []string{"From scratch", "From global", "From existing"}
	descriptions := []string{
		"Start clean. The new entry has no region override, no saved scope, and no stored token; all values fall through to the global profile at read time.",
		"Same on-disk result as 'From scratch' today (no copy) - the new entry simply inherits global values when read. Use this when the global defaults are already what you want for this profile.",
		"Copy non-sensitive overrides (region, scope project, scope service) from another repo profile in this file. Stored secrets (token, VPN keys) are never copied - each profile keeps its own.",
	}
	idx, err := uxBlock.Run(
		prompt.NewRoot(ctx, "What should the new profile start from?", choices, prompt.WithChoiceDescriptions(descriptions)),
		prompt.GetChoiceCursor,
	)
	if err != nil {
		return 0, "", err
	}

	switch profileSource(idx) {
	case profileSourceScratch:
		return profileSourceScratch, "", nil
	case profileSourceGlobal:
		return profileSourceGlobal, "", nil
	case profileSourceExisting:
		name, err := promptExistingProfileName(ctx, cmdData)
		if err != nil {
			return 0, "", err
		}
		return profileSourceExisting, name, nil
	}
	return 0, "", errors.New("unexpected source selection")
}

// promptExistingProfileName shows a list of existing repo profiles and returns the picked one.
// Errors when there are no other profiles to choose from (the current repo has only the entry
// we'd be cloning from itself, which would be a no-op copy).
func promptExistingProfileName(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) (string, error) {
	all := cmdData.Profile.RepoProfiles()
	if len(all) == 0 {
		return "", errors.New("no existing repo profiles to copy from")
	}
	rows := table.NewBody()
	for _, p := range all {
		rows.AddStringsRow(p.Name, p.Id)
	}
	header := table.NewRowFromStrings("Name", "Id")
	choice, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			rows,
			selector.WithLabel("Pick a profile to copy from"),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return "", err
	}
	if choice < 0 || choice >= len(all) {
		return "", errors.New("invalid profile selection")
	}
	return all[choice].Name, nil
}

func profileSourceLabel(source profileSource, sourceName string) string {
	switch source {
	case profileSourceScratch:
		return "scratch (empty, falls through to global)"
	case profileSourceGlobal:
		return "global (inherits at read time, no copy on disk)"
	case profileSourceExisting:
		return "existing profile \"" + sourceName + "\""
	}
	return "unknown"
}

func profileRemoveCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("remove").
		Aliases("delete").
		Short("Remove an inactive named profile from the current repo (and wipe its per-profile secrets).").
		HelpFlag("Help for the 'profile remove' command.").
		Arg("name").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			name := cmdData.Args["name"][0]
			if err := cmdData.Profile.RemoveRepoProfile(name); err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Removed profile", name))
			return nil
		})
}

func profileRenameCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("rename").
		Short("Rename an existing profile in the current repo (id and stored secrets are preserved).").
		HelpFlag("Help for the 'profile rename' command.").
		Arg("old-name").
		Arg("new-name").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			oldName := cmdData.Args["old-name"][0]
			newName := cmdData.Args["new-name"][0]
			if err := cmdData.Profile.RenameRepoProfile(oldName, newName); err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Renamed profile", oldName+" → "+newName))
			return nil
		})
}

func profileUseCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("use").
		Short("Switch the active repo profile (persisted in .zcli/profile.json).").
		HelpFlag("Help for the 'profile use' command.").
		Arg("name").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			name := cmdData.Args["name"][0]
			if err := cmdData.Profile.UseRepoProfile(name); err != nil {
				return err
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Active profile", name))
			return nil
		})
}

func profileWhereCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("where").
		Short("Show which profile layer each effective value comes from.").
		HelpFlag("Help for the 'profile where' command.").
		GuestRunFunc(func(_ context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			return printProfileWhere(cmdData)
		})
}

func printProfileWhere(cmdData *cmdBuilder.GuestCmdData) error {
	mgr := cmdData.Profile
	out := cmdData.Stdout

	out.Println("Storage")
	out.Println(formatRow("  global profile", mgr.GlobalPath()))
	if repoDir := mgr.RepoDir(); repoDir != "" {
		repoProfile, _ := mgr.RepoProfile()
		out.Println(formatRow("  repo profile", profile.RepoProfilePath(repoDir)))
		out.Println(formatRow("  active profile", repoProfile.Name+"  (id "+repoProfile.Id+")"))
	} else {
		out.Println(formatRow("  repo profile", "(none — run `zcli profile init` inside a repo to create one)"))
	}
	out.Println(formatRow("  secrets", mgr.SecretsDescription()))
	out.Println("")

	out.Println("Effective values")
	out.Println(formatRow("  region", regionLine(mgr)))
	out.Println(formatRow("  token", tokenLine(mgr)))
	out.Println(formatRow("  scope project", scopeLine(mgr)))
	out.Println(formatRow("  scope service", scopeServiceLine(mgr)))
	return nil
}

func regionLine(mgr *profile.Manager) string {
	effective := mgr.Region()
	repo, hasRepo := mgr.RepoProfile()
	if hasRepo && repo.Region != nil {
		return effective.Name + "  (from repo override)"
	}
	if effective.Name == "" {
		return "(not set)"
	}
	return effective.Name + "  (from global)"
}

func tokenLine(mgr *profile.Manager) string {
	_, src, err := mgr.TokenWithSource()
	if err != nil {
		return "(error reading: " + err.Error() + ")"
	}
	switch src {
	case profile.SourceRepo:
		return "set  (from repo override)"
	case profile.SourceGlobal:
		return "set  (from global)"
	case profile.SourceNone:
		return "(not set)"
	}
	return "(not set)"
}

func scopeLine(mgr *profile.Manager) string {
	id, ok := mgr.ScopeProject()
	if !ok {
		return "(not set)"
	}
	return id.Native() + "  (from repo)"
}

func scopeServiceLine(mgr *profile.Manager) string {
	id, ok := mgr.ScopeService()
	if !ok {
		return "(not set)"
	}
	return id.Native() + "  (from repo)"
}

func formatRow(label, value string) string {
	const labelWidth = 22
	pad := labelWidth - len(label)
	if pad < 1 {
		pad = 1
	}
	return label + strings.Repeat(" ", pad) + value
}

// ensureGitignoreEntry appends entry to <dir>/.gitignore if not already present (matched as a
// whole line, ignoring leading/trailing whitespace). Returns true when the file was modified.
// Creates .gitignore when absent. Best-effort: callers should treat errors as a warning, not
// a hard failure - the profile itself is already written by the time this runs.
func ensureGitignoreEntry(dir, entry string) (bool, error) {
	path := filepath.Join(dir, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	for _, line := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(line) == entry {
			return false, nil
		}
	}
	var out strings.Builder
	out.Write(existing)
	// Avoid sticking the new entry onto the end of a non-newline-terminated line.
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		out.WriteByte('\n')
	}
	out.WriteString(entry)
	out.WriteByte('\n')
	if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
