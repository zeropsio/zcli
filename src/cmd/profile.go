package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

func profileCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("profile").
		Short("Manage per-repository CLI profiles.").
		HelpFlag("Help for the 'profile' command.").
		AddChildrenCmd(profileInitCmd())
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
