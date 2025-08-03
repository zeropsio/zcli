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
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func projectScopeCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("scope").
		Short(i18n.T(i18n.CmdDescProjectScope)).
		ScopeLevel(cmdBuilder.ScopeProject(cmdBuilder.WithSkipSelectProject())).
		Arg(cmdBuilder.ProjectArgName, cmdBuilder.OptionalArg()).
		HelpFlag(i18n.T(i18n.CmdHelpProjectScope)).
		BoolFlag("clear", false, "Clear project scope").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			var project entity.Project
			clearProject := cmdData.Params.GetBool("clear")
			if !clearProject {
				var projectId string
				if projectIds, hasArgs := cmdData.GuestCmdData.Args["project-id"]; hasArgs && len(projectIds) > 0 {
					projectId = projectIds[0]
				}

				selectedProject, selected, err := cmdData.ProjectSelector(ctx, cmdData, uxHelpers.WithPreselectedProjectId(projectId))
				if err != nil {
					return err
				}
				if !selected {
					return errors.New("project not selected")
				}
				project = selectedProject

				question := styles.NewStringBuilder()
				question.WriteString("Project ")
				question.WriteStyledString(
					styles.SelectStyle().
						Bold(true),
					project.Name.String(),
				)
			}

			localZCliYamlFileName, exists := cmdData.Params.GetLocalZCliYamlFileName()
			if !exists {
				localZCliYamlFileName = constants.CliZcliYamlBaseFileName + ".yml"
			}

			zcliFileName := localZCliYamlFileName
			zcliFileNameNew := localZCliYamlFileName + ".tmp"
			defer os.Remove(zcliFileNameNew)

			if err := func() error {
				tmpFile, err := os.Create(zcliFileNameNew)
				if err != nil {
					return err
				}
				defer tmpFile.Close()

				written, err := func() (written bool, err error) {
					readFile, err := os.Open(zcliFileName)
					if err != nil {
						if os.IsNotExist(err) {
							return false, nil
						}
						return false, err
					}
					defer readFile.Close()

					scanner := bufio.NewScanner(readFile)
					for scanner.Scan() {
						if strings.HasPrefix(scanner.Text(), "projectId:") {
							if clearProject {
								continue
							}
							fmt.Fprintf(tmpFile, "# set by zcli project scope %s\n", project.Name.Native())
							fmt.Fprintf(tmpFile, "projectId: %s\n", project.Id.Native())
							written = true
							continue
						}
						if strings.HasPrefix(scanner.Text(), "# set by zcli project scope ") {
							continue
						}
						fmt.Fprintln(tmpFile, scanner.Text())
					}
					return written, nil
				}()

				if err != nil {
					return err
				}
				if written {
					return nil
				}
				if !clearProject {
					fmt.Fprintf(tmpFile, "# set by zcli project scope %s\n", project.Name.Native())
					fmt.Fprintf(tmpFile, "projectId: %s\n", project.Id.Native())
				}

				return nil
			}(); err != nil {
				return err
			}

			if err := os.Remove(zcliFileName); err != nil {
				return err
			}
			if err := os.Rename(zcliFileNameNew, zcliFileName); err != nil {
				return err
			}

			return nil
		})
}
