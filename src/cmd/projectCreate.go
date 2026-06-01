package cmd

import (
	"context"
	"slices"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func formatAllowedTemplateFields(fields []string) string {
	return strings.Join(fields, ", ")
}

func projectCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short("Crates a new project for specified org.").
		StringFlag("name", "", "Project name").
		StringFlag("org-id", "", "Organization ID to create project for").
		StringSliceFlag("tags", nil, "Project tags. Comma separated list or repeated flag.").
		StringFlag("out", "", "Output format of command, using golang's text/template engine. Entity fields: "+formatAllowedTemplateFields(entity.ProjectFields)).
		StringFlag("mode", enumDefaultForFlag(entity.ProjectModeLight), "Project mode"+enumValuesForFlag(entity.AllPublicProjectModes())).
		StringFlag("env-isolation", "service", "Env isolation rule [service, none] for more info see docs https://docs.zerops.io/features/env-variables#isolation-modes").
		StringFlag("ssh-isolation", "vpn", "SSH isolation rules, for more info see docs https://docs.zerops.io/references/ssh#ssh-access-control").
		StringFlag("location", "", "Project primary/fallback location (e.g. eu-central).").
		HelpFlag("Help for the project create command.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			var err error

			mode := entity.ProjectMode(strings.ToUpper(cmdData.Params.GetString("mode")))
			if !slices.Contains(entity.AllPublicProjectModes(), mode) {
				return errors.Errorf("Invalid --mode, expected one of %v, got %s", entity.AllPublicProjectModes(), mode)
			}

			outFormat := cmdData.Params.GetString("out")
			var outTemplate *template.Template
			if outFormat != "" {
				outTemplate, err = template.New("out").Parse(outFormat)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			org, err := resolveScopedOrg(ctx, cmdData)
			if err != nil {
				return err
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected org", org.Name))

			label := styles.NewStringBuilder()
			label.WriteString("Type ")
			label.WriteStyledString(
				styles.SelectStyle().
					Bold(true),
				"project",
			)
			label.WriteString(" name")

			name := cmdData.Params.GetString("name")
			if name == "" && terminal.IsTerminal() {
				name, err = uxBlock.Run(
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
					return err
				}
			} else if name == "" {
				return errors.New("Must specify name with --name")
			}

			postProject := entity.PostProject{
				OrgId:        org.Id,
				Name:         name,
				Tags:         cmdData.Params.GetStringSlice("tags"),
				Mode:         mode,
				SshIsolation: cmdData.Params.GetString("ssh-isolation"),
				EnvIsolation: cmdData.Params.GetString("env-isolation"),
			}
			if loc := cmdData.Params.GetString("location"); loc != "" {
				postProject.Location = optional.New(entity.LocationId(loc))
			} else if terminal.IsTerminal() {
				location, err := uxHelpers.PrintLocationSelector(ctx, cmdData.RestApiClient)
				if err != nil {
					return err
				}
				postProject.Location = optional.New(location.Id)
				cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected location", location.Name))
			}

			project, err := repository.PostProject(ctx, cmdData.RestApiClient, postProject)
			if err != nil {
				return err
			}

			cmdData.UxBlocks.PrintSuccessText("Project created")

			if outTemplate != nil {
				if err := outTemplate.Execute(cmdData.Stdout, project); err != nil {
					return errors.WithStack(err)
				}
			}

			return nil
		})
}
