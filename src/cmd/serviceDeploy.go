package cmd

import (
	"context"
	"io"
	"sync"

	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
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
		BoolFlag("verbose", false, i18n.T(i18n.VerboseFlag), cmdBuilder.ShortHand("v")).
		BoolFlag("deployGitFolder", false, i18n.T(i18n.UploadGitFolder), cmdBuilder.ShortHand("g")).
		HelpFlag(i18n.T(i18n.CmdHelpServiceDeploy)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			arch := archiveClient.New(archiveClient.Config{
				Logger:          uxBlocks.GetDebugFileLogger(),
				Verbose:         cmdData.Params.GetBool("verbose"),
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
					F: func(ctx context.Context, _ *uxHelpers.Process) error {
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

						reader, writer := io.Pipe()
						var finalReader io.Reader = reader
						if cmdData.Params.GetString("archiveFilePath") != "" {
							packageFile, err := openPackageFile(
								cmdData.Params.GetString("archiveFilePath"),
								cmdData.Params.GetString("workingDir"),
							)
							if err != nil {
								return err
							}
							if _, err := packageFile.Stat(); err != nil {
								return err
							}

							finalReader = io.TeeReader(reader, packageFile)
						}

						wg := sync.WaitGroup{}
						wg.Add(1)
						go func() {
							defer wg.Done()
							err := arch.TarFiles(writer, files)
							writer.CloseWithError(err)
						}()

						if err := packageStream(ctx, cmdData.RestApiClient, appVersion.Id, finalReader); err != nil {
							// if an error occurred while packing the app, return that error
							return err
						}

						// Wait for upload to finish
						wg.Wait()

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
