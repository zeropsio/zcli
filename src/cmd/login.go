package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/cliAction/login"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/grpcApiClientFactory"
	"github.com/zerops-io/zcli/src/grpcDaemonClientFactory"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
)

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "login {token | username password}",
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
				HttpTimeout: time.Second * 5,
			})

			apiClientFactory := grpcApiClientFactory.New(grpcApiClientFactory.Config{
				CaCertificateUrl: params.GetPersistentString(constants.PersistentParamCaCertificateUrl),
			})

			email, password, token := getCredentials(cmd, args)

			return login.New(
				login.Config{
					RestApiAddress: params.GetPersistentString(constants.PersistentParamRestApiAddress),
					GrpcApiAddress: params.GetPersistentString(constants.PersistentParamGrpcApiAddress),
				},
				storage,
				httpClient,
				apiClientFactory,
				grpcDaemonClientFactory.New(),
			).Run(ctx, login.RunConfig{
				ZeropsEmail:    email,
				ZeropsPassword: password,
				ZeropsToken:    token,
			})
		},
	}

	params.RegisterString(cmd, "zeropsLogin", "", "zerops account login")
	params.RegisterString(cmd, "zeropsPassword", "", "zerops account password")
	params.RegisterString(cmd, "zeropsToken", "", "zerops account token")

	return cmd
}

func getCredentials(cmd *cobra.Command, args []string) (login, password, token string) {
	login = params.GetString(cmd, "zeropsLogin")
	password = params.GetString(cmd, "zeropsPassword")
	token = params.GetString(cmd, "zeropsToken")
	if len(args) == 2 {
		login = args[0]
		password = args[1]
		token = ""
	}
	if len(args) == 1 {
		token = args[0]
		login = ""
		password = ""
	}
	return
}
