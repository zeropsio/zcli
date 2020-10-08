package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/cliAction/login"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
)

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "login",
		Short:        i18n.CmdLogin,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}

			httpClient := httpClient.New(httpClient.Config{
				HttpTimeout: time.Second * 10,
			})

			return login.New(
				login.Config{
					ApiAddress: params.GetString("restApiAddress"),
				},
				storage,
				httpClient,
			).Run(ctx, login.RunConfig{
				ZeropsLogin:    params.GetString("zeropsLogin"),
				ZeropsPassword: params.GetString("zeropsPassword"),
			})
		},
	}

	params.RegisterString(cmd, "zeropsLogin", "", "zerops account login")
	params.RegisterString(cmd, "zeropsPassword", "", "zerops account password")

	return cmd
}
