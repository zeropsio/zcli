package cmd

import (
	"context"
	"encoding/base64"
	"io"
	"time"

	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

func servicePushCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("push").
		Short(i18n.T(i18n.CmdPushDesc)).
		Long(i18n.T(i18n.CmdPushDesc)+"\n\n"+i18n.T(i18n.PushDescLong)).
		ScopeLevel(cmdBuilder.Service).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("archiveFilePath", "", i18n.T(i18n.BuildArchiveFilePath)).
		StringFlag("versionName", "", i18n.T(i18n.BuildVersionName)).
		StringFlag("source", "", i18n.T(i18n.SourceName)).
		StringFlag("zeropsYamlPath", "", i18n.T(i18n.ZeropsYamlLocation)).
		BoolFlag("deployGitFolder", false, i18n.T(i18n.UploadGitFolder)).
		HelpFlag(i18n.T(i18n.ServicePushHelp)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			arch := archiveClient.New(archiveClient.Config{
				DeployGitFolder: cmdData.Params.GetBool("deployGitFolder"),
			})

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.BuildDeployCreatingPackageStart)))

			configContent, err := getValidConfigContent(
				uxBlocks,
				cmdData.Params.GetString("workingDir"),
				cmdData.Params.GetString("zeropsYamlPath"),
			)
			if err != nil {
				return err
			}

			err = validateZeropsYamlContent(ctx, cmdData.RestApiClient, cmdData.Service, configContent)
			if err != nil {
				return err
			}

			files, err := arch.FindGitFiles(cmdData.Params.GetString("workingDir"))
			if err != nil {
				return err
			}

			var reader io.Reader
			pipeReader, writer := io.Pipe()
			defer pipeReader.Close()
			reader = pipeReader

			tarErrChan := make(chan error, 1)

			go arch.TarFiles(writer, files, tarErrChan)

			if cmdData.Params.GetString("archiveFilePath") != "" {
				packageFile, err := openPackageFile(
					cmdData.Params.GetString("archiveFilePath"),
					cmdData.Params.GetString("workingDir"),
				)
				if err != nil {
					return err
				}
				reader = io.TeeReader(reader, packageFile)
			}

			appVersion, err := createAppVersion(
				ctx,
				cmdData.RestApiClient,
				cmdData.Service,
				cmdData.Params.GetString("versionName"),
			)
			if err != nil {
				return err
			}

			// TODO - janhajek merge with sdk client
			httpClient := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F: func(ctx context.Context) error {
						if err := packageUpload(ctx, httpClient, appVersion.UploadUrl.String(), reader); err != nil {
							// if an error occurred while packing the app, return that error
							select {
							case err := <-tarErrChan:
								return err
							default:
								return err
							}
						}

						// wait for packing and saving to finish (should already be done after the package upload has finished)
						if tarErr := <-tarErrChan; tarErr != nil {
							return tarErr
						}

						return nil
					},
					RunningMessage:      i18n.T(i18n.BuildDeployUploadingPackageStart),
					ErrorMessageMessage: i18n.T(i18n.BuildDeployUploadPackageFailed),
					SuccessMessage:      i18n.T(i18n.BuildDeployUploadingPackageDone),
				}},
			)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.BuildDeployCreatingPackageDone)))

			if cmdData.Params.GetString("archiveFilePath") != "" {
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.BuildDeployPackageSavedInto, cmdData.Params.GetString("archiveFilePath"))))
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.BuildDeployDeployingStart)))

			sourceName := cmdData.Params.GetString("source")
			if sourceName == "" {
				sourceName = cmdData.Service.Name.String()
			}

			deployResponse, err := cmdData.RestApiClient.PutAppVersionBuildAndDeploy(ctx,
				path.AppVersionId{
					Id: appVersion.Id,
				},
				body.PutAppVersionBuildAndDeploy{
					BuildConfigContent: types.MediumText(base64.StdEncoding.EncodeToString(configContent)),
					Source:             types.NewStringNull(sourceName),
				},
			)
			if err != nil {
				return err
			}

			deployProcess, err := deployResponse.Output()
			if err != nil {
				return err
			}

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F:                   uxHelpers.CheckZeropsProcess(deployProcess.Id, cmdData.RestApiClient),
					RunningMessage:      i18n.T(i18n.PushRunning),
					ErrorMessageMessage: i18n.T(i18n.PushRunning),
					SuccessMessage:      i18n.T(i18n.PushFinished),
				}},
			)
			if err != nil {
				return err
			}

			return nil
		})
}
