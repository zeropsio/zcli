package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceEnvCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("env").
		Short(i18n.T(i18n.CmdDescServiceEnv)).
		ScopeLevel(scope.Service).
		Arg(scope.ServiceArgName, cmdBuilder.OptionalArg()).
		StringFlag("format", "env", i18n.T(i18n.ServiceEnvFormatFlag)).
		BoolFlag("no-secrets", false, i18n.T(i18n.ServiceEnvNoSecretsFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpServiceEnv)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			var userDataSetup repository.GetUserDataSetup
			if cmdData.Params.GetBool("no-secrets") {
				userDataSetup.EsFilters(func(filter body.EsFilter) body.EsFilter {
					filter.Search = append(filter.Search, body.EsSearchItem{
						Name:     "type",
						Operator: "ne",
						Value:    types.String(enum.UserDataTypeEnumSecret),
					})
					return filter
				})
			}

			userData, err := repository.GetUserDataByProjectId(
				ctx,
				cmdData.RestApiClient,
				cmdData.Project,
				userDataSetup,
			)
			if err != nil {
				return err
			}

			allEnvs := make(map[string]entity.UserData, len(userData))
			for _, env := range userData {
				allEnvs[fmt.Sprintf("%s_%s", env.ServiceName, env.Key)] = env
			}

			envs := make(map[string]entity.UserData)
			for key, env := range allEnvs {
				c := env.Content.Native()
				c = os.Expand(c, func(s string) string {
					e, ok := allEnvs[s]
					if ok {
						return e.Content.Native()
					}
					return s
				})
				env.Content = types.NewText(c)
				if env.ServiceId == cmdData.Service.ID {
					envs[key] = env
				}
			}

			format := cmdData.Params.GetString("format")
			formatSplit := strings.SplitN(format, "=", 2)
			formatKind := formatSplit[0]

			switch formatKind {
			case "json":
				enc := json.NewEncoder(cmdData.Stdout)
				enc.SetIndent("", "\t")
				out := make(map[string]string, len(envs))
				for _, e := range envs {
					out[e.Key.Native()] = e.Content.Native()
				}
				if err := enc.Encode(out); err != nil {
					return err
				}
			case "yaml":
				enc := yaml.NewEncoder(cmdData.Stdout)
				out := make(map[string]string, len(envs))
				for _, e := range envs {
					out[e.Key.Native()] = e.Content.Native()
				}
				if err := enc.Encode(out); err != nil {
					return err
				}
			case "value":
				for _, env := range envs {
					cmdData.Stdout.Println(env.Content)
				}
			case "go-template":
				if len(formatSplit) < 2 {
					return errors.New(i18n.T(i18n.ServiceEnvNoTemplateData))
				}
				formatTemplate := formatSplit[1]
				t, err := template.New("go").Parse(formatTemplate + "\n")
				if err != nil {
					return err
				}
				for _, value := range envs {
					if err := t.Execute(cmdData.Stdout, value); err != nil {
						return err
					}
				}
			case "env":
				for _, env := range envs {
					cmdData.Stdout.Printf("%s=%s\n", env.Key, env.Content)
				}
			default:
				return errors.New(i18n.T(i18n.ServiceEnvInvalidFormatKind, formatKind))
			}

			return nil
		})
}
