package cmd

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
)

const ZeropsYamlFileName = "zerops.yml"

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

func packageUpload(ctx context.Context, client *httpClient.Handler, uploadUrl string, reader io.Reader, options ...httpClient.Option) error {
	options = append(options, httpClient.ContentType("application/gzip"))
	cephResponse, err := client.PutStream(ctx, uploadUrl, reader, options...)
	if err != nil {
		return err
	}
	if cephResponse.StatusCode != http.StatusOK {
		return errors.New(i18n.T(i18n.BuildDeployUploadPackageFailed))
	}

	return nil
}

func getValidConfigContent(uxBlocks uxBlock.UxBlocks, workingDir string, zeropsYamlPath string) ([]byte, error) {
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
			return nil, errors.New(i18n.T(i18n.BuildDeployZeropsYamlNotFound, zeropsYamlPath))
		}
		return nil, err
	}

	uxBlocks.PrintInfo(styles.InfoLine(i18n.T(i18n.BuildDeployZeropsYamlFound, zeropsYamlPath)))

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
		ServiceStackTypeVersionName: service.ServiceStackTypeVersionName,
		ServiceStackName:            service.Name,
		ServiceStackTypeId:          service.ServiceTypeId,
		ZeropsYaml:                  types.NewMediumText(string(yamlContent)),
	})
	if err != nil {
		return err
	}
	if _, err = resp.Output(); err != nil {
		return err
	}

	return nil
}
