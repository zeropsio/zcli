package cmd

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"

	"github.com/zerops-io/zcli/src/cliAction/buildDeploy"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/archiveClient"
	"github.com/zerops-io/zcli/src/utils/httpClient"

	"github.com/spf13/cobra"
)

func pushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "push projectNameOrId serviceName",
		Short:        i18n.CmdPushDesc,
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			regSignals(cancel)

			storage, err := createCliStorage()
			if err != nil {
				return err
			}
			token := getToken(storage)

			region, err := createRegionRetriever(ctx)
			if err != nil {
				return err
			}

			reg, err := region.RetrieveFromFile()
			if err != nil {
				return err
			}

			apiClientFactory := business.New(business.Config{
				CaCertificateUrl: reg.CaCertificateUrl,
			})
			apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(
				ctx,
				reg.GrpcApiAddress,
				token,
			)
			if err != nil {
				return err
			}
			defer closeFunc()

			client := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			arch := archiveClient.New(archiveClient.Config{
				DeployGitFolder: params.GetBool(cmd, "deployGitFolder"),
			})

			return buildDeploy.New(
				buildDeploy.Config{},
				client,
				arch,
				apiGrpcClient,
				sdkConfig.Config{Token: token, RegionUrl: reg.RestApiAddress},
			).Push(ctx, buildDeploy.RunConfig{
				ArchiveFilePath:  params.GetString(cmd, "archiveFilePath"),
				WorkingDir:       params.GetString(cmd, "workingDir"),
				VersionName:      params.GetString(cmd, "versionName"),
				SourceName:       params.GetString(cmd, "source"),
				ProjectNameOrId:  args[0],
				ServiceStackName: args[1],
			})
		},
	}

	params.RegisterString(cmd, "workingDir", "./", i18n.BuildWorkingDir)
	params.RegisterString(cmd, "archiveFilePath", "", i18n.BuildArchiveFilePath)
	params.RegisterString(cmd, "versionName", "", i18n.BuildVersionName)
	params.RegisterString(cmd, "source", "", i18n.SourceName)
	params.RegisterBool(cmd, "deployGitFolder", false, i18n.ZeropsYamlLocation)

	return cmd
}
