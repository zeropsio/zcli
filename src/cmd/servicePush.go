package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmd/scope"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/dto/output"
)

func servicePushCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("push").
		Short(i18n.T(i18n.CmdDescPush)).
		Long(i18n.T(i18n.CmdDescPushLong)).
		ScopeLevel(scope.Service).
		Arg(scope.ServiceArgName, cmdBuilder.OptionalArg()).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("archiveFilePath", "", i18n.T(i18n.BuildArchiveFilePath)).
		StringFlag("versionName", "", i18n.T(i18n.BuildVersionName)).
		StringFlag("zeropsYamlPath", "", i18n.T(i18n.ZeropsYamlLocation)).
		StringFlag("setup", "", i18n.T(i18n.ZeropsYamlSetup)).
		BoolFlag("deployGitFolder", false, i18n.T(i18n.UploadGitFolder)).
		HelpFlag(i18n.T(i18n.CmdHelpPush)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {

			//			uxBlocks := cmdData.UxBlocks

			type buildLogSetup struct {
				process    *uxHelpers.Process
				appVersion *output.AppVersionJsonObject
			}
			buildLogs := make(chan buildLogSetup)
			defer close(buildLogs)
			go func() {
				setup, opened := <-buildLogs
				if !opened {
					return
				}
				fmt.Fprintf(setup.process.ViewPortWriter, "cosi")
				return
				/*
					serviceId, _ := setup.appVersion.Build.ServiceStackId.Get()
					handler := serviceLogs.New(
						serviceLogs.Config{},
						cmdData.RestApiClient,
					)
					if err := handler.Run(ctx, serviceLogs.RunConfig{
						Project:        *cmdData.Project,
						ServiceId:      serviceId,
						Limit:          100,
						MinSeverity:    "DEBUG",
						MsgType:        "APPLICATION",
						Format:         "FULL",
						FormatTemplate: "{{.message}}",
						Follow:         true,
						Tags: []string{
							"zbuilder@" + setup.appVersion.Id.Native(),
						},
						// TODO - janhajek better place?
						Levels: serviceLogs.Levels{
							{"EMERGENCY", "0"},
							{"ALERT", "1"},
							{"CRITICAL", "2"},
							{"ERROR", "3"},
							{"WARNING", "4"},
							{"NOTICE", "5"},
							{"INFORMATIONAL", "6"},
							{"DEBUG", "7"},
						},
					}); err != nil {
						panic(err)
					}

				*/
			}()
			return uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F: uxHelpers.CheckZeropsProcess("UZIp6nKAR1aHajYVAfz0fg", cmdData.RestApiClient,
						uxHelpers.CheckZeropsProcessWithProcessOutputCallback(func(process *uxHelpers.Process, apiProcess output.Process) {
							if _, exists := apiProcess.AppVersion.Build.ServiceStackId.Get(); exists {
								select {
								case buildLogs <- buildLogSetup{
									appVersion: apiProcess.AppVersion,
									process:    process,
								}:
								default:
								}
							}
						}),
					),
					RunningMessage:      i18n.T(i18n.PushRunning),
					ErrorMessageMessage: i18n.T(i18n.PushFailed),
					SuccessMessage:      i18n.T(i18n.PushFinished),
				}},
			)

			/*

				arch := archiveClient.New(archiveClient.Config{
					DeployGitFolder: cmdData.Params.GetBool("deployGitFolder"),
				})

				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployCreatingPackageStart)))

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
						F: func(ctx context.Context) (err error) {
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
								files, err := arch.FindGitFiles(ctx, cmdData.Params.GetString("workingDir"))
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

							// TODO - janhajek merge with sdk client
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

				deployResponse, err := cmdData.RestApiClient.PutAppVersionBuildAndDeploy(ctx,
					dtoPath.AppVersionId{
						Id: appVersion.Id,
					},
					body.PutAppVersionBuildAndDeploy{
						ZeropsYaml:      types.MediumText(configContent),
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

				buildLogs := make(chan *output.AppVersionJsonObject)
				defer close(buildLogs)
				go func() {
					appVersion := <-buildLogs
					serviceId, _ := appVersion.Build.ServiceStackId.Get()
					handler := serviceLogs.New(
						serviceLogs.Config{},
						cmdData.RestApiClient,
					)

					if err := handler.Run(ctx, serviceLogs.RunConfig{
						Project:        *cmdData.Project,
						ServiceId:      serviceId,
						Limit:          100,
						MinSeverity:    "DEBUG",
						MsgType:        "APPLICATION",
						Format:         "FULL",
						FormatTemplate: "{{.message}}",
						Follow:         true,
						Tags: []string{
							"zbuilder@" + appVersion.Id.Native(),
						},
						// TODO - janhajek better place?
						Levels: serviceLogs.Levels{
							{"EMERGENCY", "0"},
							{"ALERT", "1"},
							{"CRITICAL", "2"},
							{"ERROR", "3"},
							{"WARNING", "4"},
							{"NOTICE", "5"},
							{"INFORMATIONAL", "6"},
							{"DEBUG", "7"},
						},
					}); err != nil {
						panic(err)
					}
				}()
				err = uxHelpers.ProcessCheckWithSpinner(
					ctx,
					cmdData.UxBlocks,
					[]uxHelpers.Process{{
						F: uxHelpers.CheckZeropsProcess(deployProcess.Id, cmdData.RestApiClient,
							uxHelpers.CheckZeropsProcessWithProcessOutputCallback(func(process output.Process) {
								if _, exists := process.AppVersion.Build.ServiceStackId.Get(); exists {
									select {
									case buildLogs <- process.AppVersion:
									default:
									}
								}
							}),
						),
						RunningMessage:      i18n.T(i18n.PushRunning),
						ErrorMessageMessage: i18n.T(i18n.PushFailed),
						SuccessMessage:      i18n.T(i18n.PushFinished),
					}},
				)
				if err != nil {
					return err
				}

				return nil

			*/
		})
}
