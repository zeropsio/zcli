package cmd

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/serviceLogs"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
)

//nolint:maintidx // TODO (lh): remove after refactoring
func servicePushCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("push").
		Short(i18n.T(i18n.CmdDescPush)).
		Long(i18n.T(i18n.CmdDescPushLong)).
		ScopeLevel(cmdBuilder.Service(
			cmdBuilder.WithCreateNewService(),
			cmdBuilder.WithProjectScopeOptions(
				cmdBuilder.WithCreateNewProject(),
			),
		)).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("archiveFilePath", "", i18n.T(i18n.BuildArchiveFilePath)).
		StringFlag("versionName", "", i18n.T(i18n.BuildVersionName)).
		StringFlag("zeropsYamlPath", "", i18n.T(i18n.ZeropsYamlLocation)).
		StringFlag("setup", "", i18n.T(i18n.ZeropsYamlSetup)).
		BoolFlag("verbose", false, i18n.T(i18n.VerboseFlag), cmdBuilder.ShortHand("v")).
		BoolFlag("deployGitFolder", false, i18n.T(i18n.UploadGitFolder), cmdBuilder.ShortHand("g")).
		StringFlag("workspaceState", archiveClient.WorkspaceAll, i18n.T(i18n.PushWorkspaceState), cmdBuilder.ShortHand("w")).
		BoolFlag("noGit", false, i18n.T(i18n.NoGit)).
		BoolFlag("disableLogs", false, "disable logs").
		HelpFlag(i18n.T(i18n.CmdHelpPush)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			project, err := cmdData.Project.Expect("project is null")
			if err != nil {
				return err
			}
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			if cmdData.Params.IsSet("noGit") && (cmdData.Params.IsSet("deployGitFolder") || cmdData.Params.IsSet("workspaceState")) {
				uxBlocks.PrintWarning(styles.WarningLine("--noGit and --deployGitFolder/--workspaceState are mutually exclusive, ignoring --deployGitFolder/--workspaceState"))
			}

			configContent, err := yamlReader.ReadZeropsYamlContent(
				uxBlocks,
				cmdData.Params.GetString("workingDir"),
				cmdData.Params.GetString("zeropsYamlPath"),
			)
			if err != nil {
				return err
			}

			setup := service.Name
			if setupParam := cmdData.Params.GetString("setup"); setupParam != "" {
				setup = types.NewString(setupParam)
			} else {
				setups, err := yamlReader.ReadZeropsYamlSetups(configContent)
				if err != nil {
					return err
				}
				if s, hasMatch := generic.FindOne(setups, generic.ExactMatch(setup.Native())); hasMatch {
					setup = types.NewString(s)
				} else {
					selectedSetup, err := uxHelpers.PrintSetupSelector(ctx, setups)
					if err != nil {
						return err
					}
					setup = types.NewString(selectedSetup)
				}
			}
			cmdData.UxBlocks.PrintInfo(styles.InfoWithValueLine("Selected setup", setup.String()))

			err = validateZeropsYamlContent(ctx, cmdData.RestApiClient, service, setup, configContent)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployCreatingPackageStart)))

			appVersion, err := createAppVersion(
				ctx,
				cmdData.RestApiClient,
				service,
				cmdData.Params.GetString("versionName"),
			)
			if err != nil {
				return err
			}

			arch := archiveClient.New(archiveClient.Config{
				Logger:             uxBlocks.GetDebugFileLogger(),
				Verbose:            cmdData.Params.GetBool("verbose"),
				DeployGitFolder:    cmdData.Params.GetBool("deployGitFolder"),
				PushWorkspaceState: cmdData.Params.GetString("workspaceState"),
				NoGit:              cmdData.Params.GetBool("noGit"),
			})

			err = uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{{
					F: func(ctx context.Context, _ *uxHelpers.Process) (err error) {
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

							finalReader = io.TeeReader(reader, packageFile)
						}

						wg := sync.WaitGroup{}
						wg.Add(1)
						var uploadErr error
						go func() {
							defer wg.Done()

							// if an error occurred during upload, return it (it could be even auth error, before upload starts)
							if err := packageStream(ctx, cmdData.RestApiClient, appVersion.Id, finalReader); err != nil {
								_ = reader.CloseWithError(err)
								uploadErr = err // in case the reader is already closed with EOF, sometimes happened with timeouts
							}
						}()

						// if an error occurred while packing the app, return that error
						if err := arch.ArchiveGitFiles(ctx, uxBlocks, cmdData.Params.GetString("workingDir"), writer); err != nil {
							_ = writer.CloseWithError(err)
							return err
						}
						_ = writer.Close()

						// Wait for upload to finish
						wg.Wait()
						return uploadErr
					},
					RunningMessage:      i18n.T(i18n.PushDeployUploadingPackageStart),
					ErrorMessageMessage: i18n.T(i18n.PushDeployUploadPackageFailed),
					SuccessMessage:      i18n.T(i18n.PushDeployUploadingPackageDone),
				}},
			)
			if err != nil {
				return err
			}

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

			var buildPhase bool
			var preparePhase bool
			var logsHandler *serviceLogs.Handler
			if err := uxHelpers.ProcessCheckWithSpinner(
				ctx,
				cmdData.UxBlocks,
				[]uxHelpers.Process{
					{
						F: uxHelpers.CheckZeropsProcess(deployProcess.Id, cmdData.RestApiClient,
							uxHelpers.CheckZeropsProcessWithProcessOutputCallback(
								func(ctx context.Context, process *uxHelpers.Process, apiProcess output.Process) error {
									if cmdData.Params.GetBool("disableLogs") {
										return nil
									}
									if !apiProcess.Status.IsRunning() {
										return nil
									}
									if logsHandler == nil {
										logsHandler = serviceLogs.New(
											process.LogView(),
											serviceLogs.Config{},
											cmdData.RestApiClient,
										)
									}
									if !buildPhase {
										buildPhase = true
										buildServiceId, _ := apiProcess.AppVersion.Build.ServiceStackId.Get()
										go func() {
											if err := logsHandler.Run(ctx, serviceLogs.RunConfig{
												Project:        project,
												ServiceId:      buildServiceId,
												Limit:          100,
												MinSeverity:    "DEBUG",
												MsgType:        "APPLICATION",
												Format:         "FULL",
												FormatTemplate: "{{.message}}",
												Follow:         true,
												Tags: []string{
													"zbuilder@" + appVersion.Id.Native(),
												},
												Levels: serviceLogs.DefaultLevels,
											}); err != nil {
												fmt.Fprintf(logsHandler.Writer(), "\nbuild logs error: %s\n", err.Error())
											}
										}()
									}
									if !preparePhase {
										if apiProcess.AppVersion.PrepareCustomRuntime == nil {
											return nil
										}
										prepareServiceId, ok := apiProcess.AppVersion.PrepareCustomRuntime.ServiceStackId.Get()
										if !ok {
											return nil
										}
										preparePhase = true
										go func() {
											if err := logsHandler.Run(ctx, serviceLogs.RunConfig{
												Project:        project,
												ServiceId:      prepareServiceId,
												Limit:          100,
												MinSeverity:    "DEBUG",
												MsgType:        "APPLICATION",
												Format:         "FULL",
												FormatTemplate: "{{.message}}",
												Follow:         true,
												Levels:         serviceLogs.DefaultLevels,
											}); err != nil {
												fmt.Fprintf(logsHandler.Writer(), "\nprepare runtime logs error: %s\n", err.Error())
											}
										}()
									}
									return nil
								},
							),
						),
						RunningMessage:      i18n.T(i18n.PushRunning),
						ErrorMessageMessage: i18n.T(i18n.PushFailed),
						SuccessMessage:      i18n.T(i18n.PushFinished),
					},
				},
			); err != nil {
				return err
			}

			return nil
		})
}
