package cmd

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

func serviceDeployCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("deploy").
		Short(i18n.T(i18n.CmdDescDeploy)).
		Long(i18n.T(i18n.CmdDescDeployLong)).
		ScopeLevel(
			cmdBuilder.ScopeService(
				cmdBuilder.WithCreateNewService(),
				cmdBuilder.WithProjectScopeOptions(
					cmdBuilder.WithCreateNewProject(),
				),
			),
		).
		Arg(cmdBuilder.ServiceArgName, cmdBuilder.OptionalArg()).
		StringFlag("working-dir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("archive-file-path", "", i18n.T(i18n.BuildArchiveFilePath)).
		StringFlag("version-name", "", i18n.T(i18n.BuildVersionName)).
		StringFlag("zerops-yaml-path", "", i18n.T(i18n.ZeropsYamlLocation)).
		StringFlag("setup", "", i18n.T(i18n.ZeropsYamlSetup)).
		StringSliceFlag("path-to-file-or-dir", []string{"."}, "path to file or directory to be deployed. Can be repeated.").
		BoolFlag("deploy-git-folder", false, i18n.T(i18n.UploadGitFolder), cmdBuilder.ShortHand("g")).
		HelpFlag(i18n.T(i18n.CmdHelpServiceDeploy)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks
			service, err := cmdData.Service.Expect("service is null")
			if err != nil {
				return err
			}

			arch := archiveClient.New(archiveClient.Config{
				Verbose:         cmdData.Params.GetBool("verbose"),
				DeployGitFolder: cmdData.Params.GetBool("deploy-git-folder"),
			})

			configContent, err := yamlReader.ReadZeropsYamlContent(
				uxBlocks,
				cmdData.Params.GetString("working-dir"),
				cmdData.Params.GetString("zerops-yaml-path"),
			)
			if err != nil {
				return err
			}

			setups, err := yamlReader.ReadZeropsYamlSetups(configContent)
			if err != nil {
				return err
			}

			// Precedence: explicit --setup flag > auto-match by service name >
			// interactive selector (TTY) / hard error (non-TTY). The flag must
			// be consulted first; otherwise a service whose name happens to
			// equal one of the yaml setups silently ignores the user's --setup.
			var setup string
			switch {
			case cmdData.Params.IsSet("setup"):
				setup = cmdData.Params.GetString("setup")
				if err := validateSetupInYaml(setup, setups); err != nil {
					return err
				}
			default:
				if match, hasMatch := gn.FindFirst(setups, gn.ExactMatch(service.Name.String())); hasMatch {
					setup = match
				} else if !terminal.IsTerminal() {
					return errors.New("Cannot find corresponding setup in zerops.yaml, please select with --setup")
				} else {
					setup, err = uxHelpers.PrintSetupSelector(ctx, setups)
					if err != nil {
						return err
					}
				}
			}
			cmdData.UxBlocks.Print(styles.InfoWithValue("Selected setup"), setup)

			if err := validateZeropsYamlContent(
				ctx,
				cmdData.RestApiClient,
				service,
				setup,
				configContent,
			); err != nil {
				return err
			}

			uxBlocks.Print(styles.InfoLine, i18n.T(i18n.PushDeployCreatingPackageStart))

			appVersion, err := createAppVersion(
				ctx,
				cmdData.RestApiClient,
				service,
				cmdData.Params.GetString("version-name"),
			)
			if err != nil {
				return err
			}

			err = cmdData.UxBlocks.WithSpinner(ctx, uxBlock.SpinnerSpec{
				Message:        i18n.T(i18n.PushDeployUploadingPackageStart),
				ErrorMessage:   i18n.T(i18n.PushDeployUploadPackageFailed),
				SuccessMessage: i18n.T(i18n.PushDeployUploadingPackageDone),
			}, func(h uxBlock.SpinnerHandle) error {
				ignorer, err := archiveClient.LoadDeployFileIgnorer(cmdData.Params.GetString("working-dir"))
				if err != nil {
					return err
				}

				files, err := arch.FindFilesByRules(
					uxBlocks,
					cmdData.Params.GetString("working-dir"),
					cmdData.Params.GetStringSlice("path-to-file-or-dir"),
					ignorer,
				)
				if err != nil {
					return err
				}

				reader, writer := io.Pipe()
				var finalReader io.Reader = reader
				if cmdData.Params.GetString("archive-file-path") != "" {
					packageFile, err := openPackageFile(
						cmdData.Params.GetString("archive-file-path"),
						cmdData.Params.GetString("working-dir"),
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
					// Determinate progress: we know the full file list
					// up-front (FindFilesByRules), so we can render
					// "N / M files" as TarFiles processes each one.
					// More meaningful than a compressed-byte counter on
					// the upload side, where total bytes aren't known.
					err := arch.TarFiles(writer, files, archiveClient.WithProgress(func(processed, total int) {
						h.SetMessage(fmt.Sprintf("%s (%d / %d files)",
							i18n.T(i18n.PushDeployUploadingPackageStart), processed, total))
					}))
					writer.CloseWithError(err)
				}()

				if err := packageStream(ctx, appVersion.UploadUrl, finalReader); err != nil {
					// if an error occurred while packing the app, return that error
					return err
				}

				// Wait for upload to finish
				wg.Wait()
				return nil
			})
			if err != nil {
				return err
			}

			uxBlocks.Print(styles.InfoLine, i18n.T(i18n.PushDeployCreatingPackageDone))

			if cmdData.Params.GetString("archive-file-path") != "" {
				uxBlocks.Print(styles.InfoLine, i18n.T(i18n.PushDeployPackageSavedInto, cmdData.Params.GetString("archive-file-path")))
			}

			uxBlocks.Print(styles.InfoLine, i18n.T(i18n.PushDeployDeployingStart))

			deployResponse, err := cmdData.RestApiClient.PutAppVersionDeploy(
				ctx,
				dtoPath.AppVersionId{
					Id: appVersion.Id,
				},
				body.PutAppVersionDeploy{
					ZeropsYaml:      types.NewMediumTextNull(string(configContent)),
					ZeropsYamlSetup: types.NewStringNull(setup),
				},
			)
			if err != nil {
				return err
			}

			deployProcess, err := deployResponse.Output()
			if err != nil {
				return err
			}

			return uxHelpers.WatchProcess(
				ctx,
				cmdData.UxBlocks,
				cmdData.RestApiClient,
				deployProcess.Id,
				uxHelpers.ProcessMessages{
					Running: i18n.T(i18n.DeployRunning),
					Error:   i18n.T(i18n.DeployFailed),
					Success: i18n.T(i18n.DeployFinished),
				},
			)
		})
}
