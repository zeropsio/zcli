package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

// ProjectScopeCmd writes (or clears) the persisted `projectId:` line in the
// local zcli.yaml file. Streams the existing file through a temp copy and
// atomically renames into place — preserving any unrelated keys plus the
// surrounding comments. Reads its hint from --project-id (the embedded
// ProjectScope flag), but it's the positional below that the cobra version
// actually documented; both are accepted.
type ProjectScopeCmd struct {
	kongCmd.ProjectScope

	ProjectArg string `arg:"" name:"project-id" optional:"" help:"Project ID to scope to."`
	Clear      bool   `name:"clear" help:"Clear project scope."`
}

func (c *ProjectScopeCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}

	var project entity.Project
	if !c.Clear {
		hint := c.ProjectScope.ProjectID
		if hint == "" {
			hint = c.ProjectArg
		}
		picked, selected, err := auth.SelectProject(ctx, uxHelpers.WithPreselectedProjectId(hint))
		if err != nil {
			return err
		}
		if !selected {
			return errors.New("project not selected")
		}
		project = picked
	}

	zcliFileName := d.LocalZcliYamlFile
	if zcliFileName == "" {
		zcliFileName = constants.CliZcliYamlBaseFileName + ".yml"
	}
	zcliFileNameNew := zcliFileName + ".tmp"
	defer os.Remove(zcliFileNameNew)

	if err := rewriteZcliYamlProjectId(zcliFileName, zcliFileNameNew, c.Clear, project); err != nil {
		return err
	}

	// Atomic replace. If the original doesn't exist, Rename creates it from .tmp.
	if err := os.Remove(zcliFileName); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(zcliFileNameNew, zcliFileName)
}

// rewriteZcliYamlProjectId copies src → dst, transforming the projectId
// directive in flight. When src is missing, dst still gets the new directive
// (unless clear is true). When src exists, lines starting with `projectId:`
// or the `# set by zcli project scope ...` marker are filtered/replaced; all
// other lines pass through.
func rewriteZcliYamlProjectId(src, dst string, clear bool, project entity.Project) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	written, err := streamWithProjectIdRewrite(src, out, clear, project)
	if err != nil {
		return err
	}
	if written || clear {
		return nil
	}
	fmt.Fprintf(out, "# set by zcli project scope %s\n", project.Name.Native())
	fmt.Fprintf(out, "projectId: %s\n", project.Id.Native())
	return nil
}

func streamWithProjectIdRewrite(src string, out *os.File, clear bool, project entity.Project) (bool, error) {
	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer in.Close()

	var written bool
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "projectId:"):
			if clear {
				continue
			}
			fmt.Fprintf(out, "# set by zcli project scope %s\n", project.Name.Native())
			fmt.Fprintf(out, "projectId: %s\n", project.Id.Native())
			written = true
		case strings.HasPrefix(line, "# set by zcli project scope "):
			// drop stale marker; the projectId line above re-writes it.
		default:
			fmt.Fprintln(out, line)
		}
	}
	return written, nil
}
