package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/types/enum"
)

func projectEnvCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("env").
		ScopeLevel(cmdBuilder.ScopeProject()).
		Short("Print project envs to stdout.").
		HelpFlag("Help for the project env command.").
		BoolFlag("export", false, "Prepends export keyword to each env in output: 'export {{.Key}}={{.Value}}'.").
		StringFlag("template", "{{.Key}}={{.Value}}", "Output template.").
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			project, err := cmdData.Project.Expect("project is nil")
			if err != nil {
				return errors.WithStack(err)
			}

			templateString := cmdData.Params.GetString("template")
			if cmdData.Params.GetBool("export") {
				templateString = "export {{.Key}}={{.Value}}"
			}

			tmpl, err := template.New("envs").Parse(templateString)
			if err != nil {
				return errors.WithStack(err)
			}

			response, err := cmdData.RestApiClient.GetProjectEnvFile(
				ctx,
				path.ProjectId{Id: project.Id},
				query.GetProjectEnvFile{
					OverrideEnvIsolation: enum.GetProjectEnvFileOverrideEnvIsolationEnumNone,
				},
			)
			if err != nil {
				return errors.WithStack(err)
			}

			envs, err := response.Output()
			if err != nil {
				return errors.WithStack(err)
			}

			output := new(strings.Builder)
			scanner := bufio.NewScanner(strings.NewReader(envs.EnvFile.String()))
			for scanner.Scan() {
				parts := strings.SplitN(scanner.Text(), "=", 2)
				if len(parts) != 2 {
					cmdData.UxBlocks.PrintWarningText(fmt.Sprintf("unexpected env format: %s", scanner.Text()))
					continue
				}
				if err := tmpl.Execute(output, Env{
					Key:   parts[0],
					Value: parts[1],
				}); err != nil {
					return errors.WithStack(err)
				}
				output.WriteRune('\n')
			}

			cmdData.Stdout.Println(output.String()[:output.Len()-1])

			return nil
		})
}

type Env struct {
	Key, Value string
}
