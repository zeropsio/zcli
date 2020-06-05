package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/command/login"
	"github.com/zerops-io/zcli/src/service/httpClient"

	"github.com/spf13/cobra"
)

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "login",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel, logger)

			storage, err := createStorage()
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
				logger,
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
