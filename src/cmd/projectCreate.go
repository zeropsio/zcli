package cmd

import (
	"context"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/input"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func projectCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short("").
		StringFlag("name", "", "Project name").
		StringFlag("orgId", "", "Organization ID to create project for").
		StringSliceFlag("tags", nil, "Project tags. Comma separated list or repeated flag.").
		StringFlag("out", "", "Output format of command").
		StringFlag("mode", strings.ToLower(enum.ProjectModeEnumLight.String()), "Project mode ['light', 'serious']").
		StringFlag("envIsolation", "service", "Env isolation setting ['service', 'none']").
		StringFlag("sshIsolation", "vpn", "SSH isolation setting, for more see docs <TODO link>").
		HelpFlag("").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			var err error

			mode := cmdData.Params.GetString("mode")
			mode = strings.ToUpper(mode)
			if !enum.ProjectModeEnum(mode).Is(enum.ProjectModeEnumAllPublic()...) {
				return errors.Errorf("Invalid --mode, expected one of %s, got %s", enum.ProjectModeEnumAllPublic(), mode)
			}

			outFormat := cmdData.Params.GetString("out")
			var outTemplate *template.Template
			if outFormat != "" {
				outTemplate, err = template.New("out").Parse(outFormat)
				if err != nil {
					return errors.WithStack(err)
				}
			}

			orgId := cmdData.Params.GetString("orgId")
			var org entity.Org
			switch {
			case orgId != "":
				org, err = repository.GetOrgById(
					ctx,
					cmdData.RestApiClient,
					uuid.ClientId(orgId),
				)
				if err != nil {
					return err
				}
			case !terminal.IsTerminal():
				return errors.New("Must specify organization ID with --orgId")
			default:
				org, err = uxHelpers.PrintOrgSelector(
					ctx,
					cmdData.RestApiClient,
					uxHelpers.WithOrgPickOnlyOneItem(true),
				)
				if err != nil {
					return err
				}
			}

			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected org", org.Name.String()))

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
				name, err = uxBlock.RunR(
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

			project, err := repository.PostProject(ctx, cmdData.RestApiClient, repository.ProjectPost{
				ClientId:     org.ID,
				Name:         types.NewString(name),
				Tags:         cmdData.Params.GetStringSlice("tags"),
				Mode:         enum.ProjectModeEnum(mode),
				SshIsolation: types.NewStringNull(cmdData.Params.GetString("sshIsolation")),
				EnvIsolation: types.NewStringNull(cmdData.Params.GetString("envIsolation")),
			})
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
