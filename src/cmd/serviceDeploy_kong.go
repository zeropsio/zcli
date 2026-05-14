package cmd

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/dto/input/body"
	dtoPath "github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

// ServiceDeployCmd packages the workspace and deploys it as a new app
// version. Uses the service scope hook with WithCreateNewService so a
// missing service can be created on the fly from zerops.yaml.
type ServiceDeployCmd struct {
	kongCmd.ServiceScope

	WorkingDir      string   `name:"working-dir" default:"./" help:"Working directory."`
	ArchiveFilePath string   `name:"archive-file-path" help:"Save the packaged archive to this path."`
	VersionName     string   `name:"version-name" help:"Name for the new app version."`
	ZeropsYamlPath  string   `name:"zerops-yaml-path" help:"Path to zerops.yaml."`
	Setup           string   `name:"setup" help:"zerops.yaml setup name."`
	PathToFileOrDir []string `name:"path-to-file-or-dir" default:"." help:"Path to file or directory to be deployed. Can be repeated."`
	Verbose         bool     `name:"verbose" short:"v" help:"Verbose logging."`
	DeployGitFolder bool     `name:"deploy-git-folder" short:"g" help:"Include .git in the deploy."`
}

func (c *ServiceDeployCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
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

	uxBlocks := d.UxBlocks
	arch := archiveClient.New(archiveClient.Config{
		Logger:          uxBlocks.GetDebugFileLogger(),
		Verbose:         c.Verbose,
		DeployGitFolder: c.DeployGitFolder,
	})

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

	err = uxHelpers.ProcessCheckWithSpinner(ctx, uxBlocks, []uxHelpers.Process{{
		F: func(ctx context.Context, _ *uxHelpers.Process) error {
			ignorer, err := archiveClient.LoadDeployFileIgnorer(c.WorkingDir)
			if err != nil {
				return err
			}
			files, err := arch.FindFilesByRules(uxBlocks, c.WorkingDir, c.PathToFileOrDir, ignorer)
			if err != nil {
				return err
			}

			reader, writer := io.Pipe()
			var finalReader io.Reader = reader
			if c.ArchiveFilePath != "" {
				packageFile, err := openPackageFile(c.ArchiveFilePath, c.WorkingDir)
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

			if err := packageStream(ctx, appVersion.UploadUrl, finalReader); err != nil {
				return err
			}
			wg.Wait()
			return nil
		},
		RunningMessage:      i18n.T(i18n.PushDeployUploadingPackageStart),
		ErrorMessageMessage: i18n.T(i18n.PushDeployUploadPackageFailed),
		SuccessMessage:      i18n.T(i18n.PushDeployUploadingPackageDone),
	}})
	if err != nil {
		return err
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployCreatingPackageDone)))
	if c.ArchiveFilePath != "" {
		uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployPackageSavedInto, c.ArchiveFilePath)))
	}
	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployDeployingStart)))

	deployResp, err := auth.RestApiClient.PutAppVersionDeploy(
		ctx,
		dtoPath.AppVersionId{Id: appVersion.Id},
		body.PutAppVersionDeploy{
			ZeropsYaml:      types.NewMediumTextNull(string(configContent)),
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

	return uxHelpers.ProcessCheckWithSpinner(ctx, uxBlocks, []uxHelpers.Process{{
		F:                   uxHelpers.CheckZeropsProcess(deployProcess.Id, auth.RestApiClient, uxBlocks),
		RunningMessage:      i18n.T(i18n.DeployRunning),
		ErrorMessageMessage: i18n.T(i18n.DeployFailed),
		SuccessMessage:      i18n.T(i18n.DeployFinished),
	}})
}
