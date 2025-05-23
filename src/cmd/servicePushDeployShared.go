package cmd

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

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

func OpenPackageFile(archiveFilePath string, workingDir string) (*os.File, error) {
	return openPackageFile(archiveFilePath, workingDir)
}

func openPackageFile(archiveFilePath string, workingDir string) (*os.File, error) {
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	archiveFilePath = filepath.Join(workingDir, archiveFilePath)

	filePath, err := filepath.Abs(archiveFilePath)
	if err != nil {
		return nil, err
	}

	// check if the target file exists
	_, err = os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		return nil, errors.Errorf(i18n.T(i18n.ArchClientFileAlreadyExists), archiveFilePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func packageStream(ctx context.Context, restApiClient *zeropsRestApiClient.Handler, appVersionId uuid.AppVersionId, reader io.Reader) error {
	// TODO(ms): content-type: application/octet-stream
	upload, err := restApiClient.PutAppVersionUpload(ctx, path.AppVersionId{Id: appVersionId}, reader)
	if err != nil {
		return err
	}
	if _, err := upload.Output(); err != nil {
		return err
	}
	if upload.StatusCode() != http.StatusOK {
		return errors.New(i18n.T(i18n.PushDeployUploadPackageFailed))
	}
	return nil
}

func getValidConfigContent(uxBlocks uxBlock.UxBlocks, selectedWorkingDir string, selectedZeropsYamlPath string) ([]byte, error) {
	workingDir, err := filepath.Abs(selectedWorkingDir)
	if err != nil {
		return nil, err
	}

	var pathsToCheck []string
	if selectedZeropsYamlPath != "" {
		if filepath.IsAbs(selectedZeropsYamlPath) {
			pathsToCheck = append(pathsToCheck, selectedZeropsYamlPath)
		} else {
			pathsToCheck = append(pathsToCheck, filepath.Join(workingDir, selectedZeropsYamlPath))
		}
	} else {
		pathsToCheck = append(pathsToCheck, filepath.Join(workingDir, "zerops.yaml"))
		pathsToCheck = append(pathsToCheck, filepath.Join(workingDir, "zerops.yml"))
	}

	zeropsYamlPath, err := func() (string, error) {
		for _, path := range pathsToCheck {
			zeropsYamlStat, err := os.Stat(path)
			if err == nil {
				uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.PushDeployZeropsYamlFound, path)))

				if zeropsYamlStat.Size() == 0 {
					return "", errors.New(i18n.T(i18n.PushDeployZeropsYamlEmpty))
				}
				if zeropsYamlStat.Size() > 10*1024 {
					return "", errors.New(i18n.T(i18n.PushDeployZeropsYamlTooLarge))
				}
				return path, nil
			}
		}
		return "", errors.New(i18n.T(i18n.PushDeployZeropsYamlNotFound, strings.Join(pathsToCheck, ", ")))
	}()
	if err != nil {
		return nil, err
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
	setup types.String,
	yamlContent []byte,
) error {
	resp, err := restApiClient.PostServiceStackZeropsYamlValidation(ctx, body.ZeropsYamlValidation{
		ServiceStackTypeVersionName: service.ServiceStackTypeVersionName,
		ServiceStackName:            service.Name,
		ServiceStackTypeId:          service.ServiceTypeId,
		ZeropsYaml:                  types.NewMediumText(string(yamlContent)),
		ZeropsYamlSetup:             setup.StringNull(),
	})
	if err != nil {
		return err
	}
	if _, err = resp.Output(); err != nil {
		return errorsx.Convert(
			err,
			errorsx.And(
				errorsx.ErrorCode("zeropsYamlServiceNotFound"),
				errorsx.Meta(func(_ apiError.Error, metaItem map[string]interface{}) string {
					if name, ok := metaItem["name"]; ok {
						return i18n.T(i18n.ErrorServiceNotFound, name)
					}
					return ""
				}),
			),
		)
	}

	return nil
}
