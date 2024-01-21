package cmd

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
)

func serviceDeployCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("deploy").
		Short(i18n.T(i18n.CmdDeployDesc)).
		Long(i18n.T(i18n.CmdDeployDesc)+"\n\n"+i18n.T(i18n.DeployDescLong)+"\n\n"+i18n.T(i18n.DeployHintPush)).
		ScopeLevel(cmdBuilder.Service).
		Arg("pathToFileOrDir", cmdBuilder.ArrayArg()).
		StringFlag("workingDir", "./", i18n.T(i18n.BuildWorkingDir)).
		StringFlag("archiveFilePath", "", i18n.T(i18n.BuildArchiveFilePath)).
		StringFlag("versionName", "", i18n.T(i18n.BuildVersionName)).
		StringFlag("zeropsYamlPath", "", i18n.T(i18n.SourceName)).
		BoolFlag("deployGitFolder", false, i18n.T(i18n.ZeropsYamlLocation)).
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

			err = validateZeropsYamlContent(ctx, cmdData.RestApiClient, cmdData.Service, configContent)
			if err != nil {
				return err
			}

			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployCreatingPackageStart))

			files, err := arch.FindFilesByRules(cmdData.Params.GetString("workingDir"), cmdData.Args["pathToFileOrDir"])
			if err != nil {
				return err
			}

			reader, writer := io.Pipe()
			defer reader.Close()

			tarErrChan := make(chan error, 1)

			go arch.TarFiles(writer, files, tarErrChan)

			r, err := savePackage(cmdData.Params.GetString("archiveFilePath"), reader)
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

			// TODO - janhajek merge with sdk client?
			httpClient := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			// TODO - janhajek spinner?
			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployUploadingPackageStart))
			if err := packageUpload(httpClient, appVersion.UploadUrl.String(), r); err != nil {
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

			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployUploadingPackageDone))

			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployDeployingStart))

			deployResponse, err := cmdData.RestApiClient.PutAppVersionDeploy(
				ctx,
				path.AppVersionId{
					Id: appVersion.Id,
				},
				body.PutAppVersionDeploy{
					ConfigContent: types.NewMediumTextNull(base64.StdEncoding.EncodeToString(configContent)),
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
				cmdData.RestApiClient,
				[]uxHelpers.Process{{
					Id:                  deployProcess.Id,
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

func getValidConfigContent(uxBlocks *uxBlock.UxBlocks, workingDir string, zeropsYamlPath string) ([]byte, error) {
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	if zeropsYamlPath != "" {
		workingDir = filepath.Join(workingDir, zeropsYamlPath)
	}

	zeropsYamlPath = filepath.Join(workingDir, ZeropsYamlFileName)

	zeropsYamlStat, err := os.Stat(zeropsYamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			if zeropsYamlPath != "" {
				return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlNotFound))
			}
		}
		return nil, nil
	}

	uxBlocks.PrintLine(i18n.T(i18n.BuildDeployZeropsYamlFound, zeropsYamlPath))

	if zeropsYamlStat.Size() == 0 {
		return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlEmpty))
	}
	if zeropsYamlStat.Size() > 10*1024 {
		return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlTooLarge))
	}

	yamlContent, err := os.ReadFile(zeropsYamlPath)
	if err != nil {
		return nil, err
	}

	return yamlContent, nil
}

func validateZeropsYamlContent(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service *entity.Service,
	yamlContent []byte,
) error {

	resp, err := restApiClient.PostServiceStackZeropsYamlValidation(ctx, body.ZeropsYamlValidation{
		Name:               service.Name,
		ServiceStackTypeId: service.ServiceTypeId,
		ZeropsYaml:         types.NewText(string(yamlContent)),
	})
	if err != nil {
		return err
	}
	if _, err = resp.Output(); err != nil {
		return err
	}

	return nil
}
