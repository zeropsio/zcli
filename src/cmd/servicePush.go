package cmd

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/archiveClient"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
)

// TODO - janhajek shared
const ZeropsYamlFileName = "zerops.yaml"

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
		BoolFlag("deployGitFolder", false, i18n.T(i18n.UploadGitFolder)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			arch := archiveClient.New(archiveClient.Config{
				DeployGitFolder: cmdData.Params.GetBool("deployGitFolder"),
			})

			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployCreatingPackageStart))

			files, err := arch.FindGitFiles(cmdData.Params.GetString("workingDir"))
			if err != nil {
				return err
			}

			configContent, err := buildConfigContent(files)
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

			// TODO - janhajek merge with sdk client
			HttpClient := httpClient.New(ctx, httpClient.Config{
				HttpTimeout: time.Minute * 15,
			})

			// TODO - janhajek spinner?
			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployUploadingPackageStart))
			if err := packageUpload(HttpClient, appVersion.UploadUrl.String(), r); err != nil {
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

			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployCreatingPackageDone))

			if cmdData.Params.GetString("archiveFilePath") != "" {
				uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployPackageSavedInto, cmdData.Params.GetString("archiveFilePath")))
			}

			uxBlocks.PrintInfoLine(i18n.T(i18n.BuildDeployDeployingStart))

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

func createAppVersion(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service *entity.Service,
	versionName string,
) (output.PostAppVersion, error) {
	appVersionResponse, err := restApiClient.PostAppVersion(
		ctx,
		body.PostAppVersion{
			ServiceStackId: service.ID,
			Name: func() types.StringNull {
				if versionName != "" {
					return types.NewStringNull(versionName)
				}
				return types.StringNull{}
			}(),
		},
	)
	if err != nil {
		return output.PostAppVersion{}, err
	}
	appVersion, err := appVersionResponse.Output()
	if err != nil {
		return output.PostAppVersion{}, err
	}

	return appVersion, nil
}

func savePackage(archiveFilePath string, reader io.Reader) (io.Reader, error) {
	if archiveFilePath == "" {
		return reader, nil
	}

	filePath, err := filepath.Abs(archiveFilePath)
	if err != nil {
		return reader, err
	}

	// check if the target file exists
	_, err = os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return reader, err
	}
	if err == nil {
		return reader, errors.Errorf(i18n.T(i18n.ArchClientFileAlreadyExists), archiveFilePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return reader, err
	}

	return io.TeeReader(reader, file), nil
}

func packageUpload(client *httpClient.Handler, uploadUrl string, reader io.Reader) error {
	cephResponse, err := client.PutStream(uploadUrl, reader, httpClient.ContentType("application/gzip"))
	if err != nil {
		return err
	}
	if cephResponse.StatusCode != http.StatusCreated {
		return errors.New(i18n.T(i18n.BuildDeployUploadPackageFailed))
	}

	return nil
}

func buildConfigContent(files []archiveClient.File) ([]byte, error) {
	for _, file := range files {
		if file.ArchivePath == ZeropsYamlFileName {
			stat, err := os.Stat(file.SourcePath)
			if err != nil {
				return nil, err
			}

			if stat.Size() == 0 {
				return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlEmpty))
			}
			if stat.Size() > 10*1024 {
				return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlTooLarge))
			}

			buildConfigContent, err := os.ReadFile(file.SourcePath)
			if err != nil {
				return nil, err
			}

			return buildConfigContent, nil
		}
	}

	return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlNotFound))
}
