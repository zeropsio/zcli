package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/i18n"
)

func serviceEnvCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("env").
		Short(i18n.T(i18n.CmdDescServiceEnv)).
		ScopeLevel(scope.Service).
		Arg(scope.ServiceArgName, cmdBuilder.OptionalArg()).
		StringFlag("name", "", i18n.T(i18n.ServiceEnvNameFlag)).
		StringFlag("format", "env", i18n.T(i18n.ServiceEnvFormatFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpServiceEnv)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {

			name := cmdData.Params.GetString("name")
			format := cmdData.Params.GetString("format")

			userDataList, err := repository.GetUserDataByServiceId(
				ctx,
				cmdData.RestApiClient,
				cmdData.Project,
				cmdData.Service.ID,
			)
			if err != nil {
				return err
			}

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("\t", "\t")
				enc.Encode(userDataList)
			case "yaml":
				enc := yaml.NewEncoder(os.Stdout)
				enc.Encode(userDataList)
			case "value":
				for _, userData := range userDataList {
					fmt.Println(userData.Content)
				}
			default:
				for _, userData := range userDataList {
					fmt.Printf("%s=%s\n", userData.Key, userData.Content)
				}
			}

			return nil
		})
}
