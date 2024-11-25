package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

func serviceDeployCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("deploy").
		Short(i18n.T(i18n.CmdDescDeploy)).
		Long(i18n.T(i18n.CmdDescDeployLong)).
		ScopeLevel(scope.Service).
		Arg("pathToFileOrDir", cmdBuilder.ArrayArg()).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("archiveFilePath", "", i18n.T(i18n.BuildArchiveFilePath)).
		StringFlag("versionName", "", i18n.T(i18n.BuildVersionName)).
		StringFlag("zeropsYamlPath", "", i18n.T(i18n.ZeropsYamlLocation)).
		StringFlag("setup", "", i18n.T(i18n.ZeropsYamlSetup)).
		BoolFlag("deployGitFolder", false, i18n.T(i18n.UploadGitFolder)).
		HelpFlag(i18n.T(i18n.CmdHelpServiceDeploy)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			arch := archiveClient.New(archiveClient.Config{
				DeployGitFolder: cmdData.Params.GetBool("deployGitFolder"),
			})

			configContent, err := getValidConfigContent(
				uxBlocks,
				cmdData.Params.GetString("workingDir"),
				cmdData.Params.GetString("zeropsYamlPath"),
			)
			if err != nil {
				return err
			}

			setup := cmdData.Service.Name
			if setupParam := cmdData.Params.GetString("setup"); setupParam != "" {
				setup = types.NewString(setupParam)
			}
			err = validateZeropsYamlContent(ctx, cmdData.RestApiClient, cmdData.Service, setup, configContent)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployCreatingPackageStart)))

			appVersion, err := createAppVersion(
				ctx,
				cmdData.RestApiClient,
				cmdData.Service,
				cmdData.Params.GetString("versionName"),
			)
			if err != nil {
				return err
			}

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F: func(ctx context.Context) error {
						var size int64
						var reader io.Reader

						if cmdData.Params.GetString("archiveFilePath") != "" {
							packageFile, err := openPackageFile(
								cmdData.Params.GetString("archiveFilePath"),
								cmdData.Params.GetString("workingDir"),
							)
							if err != nil {
								return err
							}
							s, err := packageFile.Stat()
							if err != nil {
								return err
							}
							size = s.Size()
							reader = packageFile
						} else {
							tempFile := filepath.Join(os.TempDir(), appVersion.Id.Native())
							f, err := os.Create(tempFile)
							if err != nil {
								return err
							}
							defer os.Remove(tempFile)
							ignorer, err := archiveClient.LoadDeployFileIgnorer(cmdData.Params.GetString("workingDir"))
							if err != nil {
								return err
							}
							files, err := arch.FindFilesByRules(
								uxBlocks,
								cmdData.Params.GetString("workingDir"),
								cmdData.Args["pathToFileOrDir"],
								ignorer,
							)
							if err != nil {
								return err
							}
							if err := arch.TarFiles(f, files); err != nil {
								return err
							}
							if err := f.Close(); err != nil {
								return err
							}
							readFile, err := os.Open(tempFile)
							if err != nil {
								return err
							}
							defer readFile.Close()
							stat, err := readFile.Stat()
							if err != nil {
								return err
							}
							size = stat.Size()
							reader = readFile
						}

						// TODO - janhajek merge with sdk client?
						client := httpClient.New(ctx, httpClient.Config{
							HttpTimeout: time.Minute * 15,
						})

						if err := packageUpload(ctx, client, appVersion.UploadUrl.String(), reader, httpClient.ContentLength(size)); err != nil {
							// if an error occurred while packing the app, return that error
							return err
						}
						return nil
					},
					RunningMessage:      i18n.T(i18n.PushDeployUploadingPackageStart),
					ErrorMessageMessage: i18n.T(i18n.PushDeployUploadPackageFailed),
					SuccessMessage:      i18n.T(i18n.PushDeployUploadingPackageDone),
				}},
			)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployCreatingPackageDone)))

			if cmdData.Params.GetString("archiveFilePath") != "" {
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployPackageSavedInto, cmdData.Params.GetString("archiveFilePath"))))
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployDeployingStart)))

			deployResponse, err := cmdData.RestApiClient.PutAppVersionDeploy(
				ctx,
				dtoPath.AppVersionId{
					Id: appVersion.Id,
				},
				body.PutAppVersionDeploy{
					ZeropsYaml:      types.NewMediumTextNull(string(configContent)),
					ZeropsYamlSetup: setup.StringNull(),
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
					RunningMessage:      i18n.T(i18n.DeployRunning),
					ErrorMessageMessage: i18n.T(i18n.DeployFailed),
					SuccessMessage:      i18n.T(i18n.DeployFinished),
				}},
			)

			if err != nil {
				return err
			}

			return nil
		})
}
