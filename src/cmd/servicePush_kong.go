package cmd

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/serviceLogs"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
)

// ServicePushCmd packages the local workspace (respecting --workspace-state /
// --no-git), uploads it as an app version, and runs build-and-deploy with
// live build/prepare logs streamed back.
type ServicePushCmd struct {
	kongCmd.ServiceScope

	WorkingDir      string `name:"working-dir" default:"./" help:"Working directory."`
	ArchiveFilePath string `name:"archive-file-path" help:"Save the packaged archive to this path."`
	VersionName     string `name:"version-name" help:"Name for the new app version."`
	ZeropsYamlPath  string `name:"zerops-yaml-path" help:"Path to zerops.yaml."`
	Setup           string `name:"setup" help:"zerops.yaml setup name."`
	Verbose         bool   `name:"verbose" short:"v" help:"Verbose logging."`
	DeployGitFolder bool   `name:"deploy-git-folder" short:"g" help:"Include .git in the deploy."`
	// WorkspaceState defaults to empty rather than archiveClient.WorkspaceAll so
	// we can distinguish "user passed --workspace-state" from "unset" — the
	// mutex check below depends on the distinction. The archive client gets
	// the resolved value with WorkspaceAll as the runtime default.
	WorkspaceState string `name:"workspace-state" short:"w" help:"Workspace state: all|clean|staged."`
	NoGit          bool   `name:"no-git" help:"Don't archive git data."`
	DisableLogs    bool   `name:"disable-logs" help:"Suppress build/prepare logs."`
}

//nolint:maintidx // ported as-is from cobra; refactor is a separate concern
func (c *ServicePushCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, err := d.RequireAuth()
	if err != nil {
		return err
	}
	service, err := auth.RequireService(ctx, c.ServiceScope,
		kongCmd.WithCreateNewService(c.WorkingDir, c.ZeropsYamlPath),
	)
	if err != nil {
		return err
	}
	project, err := auth.RequireProject(ctx, c.ProjectScope)
	if err != nil {
		return err
	}
	uxBlocks := d.UxBlocks

	if c.NoGit && (c.DeployGitFolder || c.WorkspaceState != "") {
		uxBlocks.PrintWarning(styles.WarningLine("--no-git and --deploy-git-folder/--workspace-state are mutually exclusive, ignoring --deploy-git-folder/--workspace-state"))
	}
	if c.WorkspaceState != "" && !gn.IsOneOf(c.WorkspaceState, archiveClient.WorkspaceAll, archiveClient.WorkspaceClean, archiveClient.WorkspaceStaged) {
		return errors.New("Invalid value for --workspace-state, please use one of: " +
			archiveClient.WorkspaceAll + ", " + archiveClient.WorkspaceClean + ", " + archiveClient.WorkspaceStaged)
	}

	configContent, err := yamlReader.ReadZeropsYamlContent(uxBlocks, c.WorkingDir, c.ZeropsYamlPath)
	if err != nil {
		return err
	}
	setups, err := yamlReader.ReadZeropsYamlSetups(configContent)
	if err != nil {
		return err
	}

	setup, hasMatch := gn.FindFirst(setups, gn.ExactMatch(service.Name.String()))
	if !hasMatch {
		setup = c.Setup
		switch {
		case !terminal.IsTerminal() && c.Setup == "":
			return errors.New("Cannot find corresponding setup in zerops.yaml, please select with --setup")
		case c.Setup == "":
			setup, err = uxHelpers.PrintSetupSelector(ctx, setups)
			if err != nil {
				return err
			}
		}
	}
	uxBlocks.PrintInfo(styles.InfoWithValueLine("Selected setup", setup))

	if err := validateZeropsYamlContent(ctx, auth.RestApiClient, service, setup, configContent); err != nil {
		return err
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployCreatingPackageStart)))

	appVersion, err := createAppVersion(ctx, auth.RestApiClient, service, c.VersionName)
	if err != nil {
		return err
	}

	workspaceState := c.WorkspaceState
	if workspaceState == "" {
		workspaceState = archiveClient.WorkspaceAll
	}
	arch := archiveClient.New(archiveClient.Config{
		Logger:             uxBlocks.GetDebugFileLogger(),
		Verbose:            c.Verbose,
		DeployGitFolder:    c.DeployGitFolder,
		PushWorkspaceState: workspaceState,
		NoGit:              c.NoGit,
	})

	err = uxHelpers.ProcessCheckWithSpinner(ctx, uxBlocks, []uxHelpers.Process{{
		F: func(ctx context.Context, _ *uxHelpers.Process) (err error) {
			reader, writer := io.Pipe()
			var finalReader io.Reader = reader
			if c.ArchiveFilePath != "" {
				packageFile, err := openPackageFile(c.ArchiveFilePath, c.WorkingDir)
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
				if err := packageStream(ctx, appVersion.UploadUrl, finalReader); err != nil {
					_ = reader.CloseWithError(err)
					uploadErr = err
				}
			}()

			if err := arch.ArchiveGitFiles(ctx, uxBlocks, c.WorkingDir, writer); err != nil {
				_ = writer.CloseWithError(err)
				return err
			}
			_ = writer.Close()
			wg.Wait()
			return uploadErr
		},
		RunningMessage:      i18n.T(i18n.PushDeployUploadingPackageStart),
		ErrorMessageMessage: i18n.T(i18n.PushDeployUploadPackageFailed),
		SuccessMessage:      i18n.T(i18n.PushDeployUploadingPackageDone),
	}})
	if err != nil {
		return err
	}

	if c.ArchiveFilePath != "" {
		uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployPackageSavedInto, c.ArchiveFilePath)))
	}
	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployDeployingStart)))

	deployResp, err := auth.RestApiClient.PutAppVersionBuildAndDeploy(ctx,
		dtoPath.AppVersionId{Id: appVersion.Id},
		body.PutAppVersionBuildAndDeploy{
			ZeropsYaml:      types.MediumText(configContent),
			ZeropsYamlSetup: types.NewStringNull(setup),
		},
	)
	if err != nil {
		return err
	}
	deployProcess, err := deployResp.Output()
	if err != nil {
		return err
	}

	guiHost := d.CliStorage.Data().RegionData.GuiAddress.OrDefault("app.zerops.io")

	var buildPhase bool
	var preparePhase bool
	var logsHandler *serviceLogs.Handler
	return uxHelpers.ProcessCheckWithSpinner(ctx, uxBlocks, []uxHelpers.Process{{
		F: uxHelpers.CheckZeropsProcess(deployProcess.Id, auth.RestApiClient, uxBlocks,
			uxHelpers.CheckZeropsProcessWithProcessOutputCallback(
				func(ctx context.Context, process *uxHelpers.Process, apiProcess output.Process) error {
					if c.DisableLogs || !apiProcess.Status.IsRunning() {
						return nil
					}
					if logsHandler == nil {
						pipelineLink := styles.NewStringBuilder()
						pipelineLink.WriteInfoColor("View full pipeline at ")
						pipelineLink.WriteStyledString(
							styles.SuccessStyle().Bold(true),
							fmt.Sprintf("https://%s/service-stack/%s/deploy/%s",
								guiHost, service.Id, apiProcess.AppVersion.Id),
						)
						logsHandler = serviceLogs.New(
							process.LogView(
								logView.WithAdditionalText(pipelineLink.String()),
								logView.WithVerticalOffset(3),
								logView.WithMaxHeight(max(30, int(float64(uxBlocks.TerminalHeight)*0.75))),
							),
							serviceLogs.Config{},
							auth.RestApiClient,
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
								FormatTemplate: "{{.Message}}",
								Follow:         true,
								Tags:           []string{"zbuilder@" + appVersion.Id.Native()},
								Levels:         serviceLogs.DefaultLevels,
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
								FormatTemplate: "{{.Message}}",
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
	}})
}
