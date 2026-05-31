package repository

import (
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// CreateAppVersion creates a new app version for the service and returns its id
// plus the URL the source package must be uploaded to.
func CreateAppVersion(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	serviceId entity.ServiceId,
	versionName string,
) (entity.AppVersionUpload, error) {
	name := types.StringNull{}
	if versionName != "" {
		name = types.NewStringNull(versionName)
	}
	response, err := restApiClient.PostServiceStackAppVersion(
		ctx,
		path.ServiceStackId{Id: uuid.ServiceStackId(serviceId)},
		body.PostAppVersion{Name: name},
	)
	if err != nil {
		return entity.AppVersionUpload{}, err
	}
	appVersion, err := response.Output()
	if err != nil {
		return entity.AppVersionUpload{}, err
	}
	return entity.AppVersionUpload{
		Id:        entity.AppVersionId(appVersion.Id),
		UploadUrl: appVersion.UploadUrl.Native(),
	}, nil
}

// UploadAppVersionPackage streams the source package to the app version's upload URL.
func UploadAppVersionPackage(ctx context.Context, uploadUrl string, reader io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadUrl, reader)
	if err != nil {
		return errors.Errorf("failed to create upload HTTP request: PUT %s => %v", uploadUrl, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Errorf("failed to upload HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return errors.New(i18n.T(i18n.PushDeployUploadPackageFailed))
	}

	return nil
}

// ValidateZeropsYaml validates the zerops.yaml content against the service before deploy.
func ValidateZeropsYaml(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	service entity.Service,
	setup string,
	yamlContent []byte,
) error {
	resp, err := restApiClient.PostServiceStackZeropsYamlValidation(ctx, body.ZeropsYamlValidation{
		ServiceStackTypeVersionName: types.NewString(service.ServiceStackTypeVersionName),
		ServiceStackName:            types.NewString(service.Name),
		ServiceStackTypeId:          stringId.ServiceStackTypeId(service.ServiceTypeId),
		ZeropsYaml:                  types.NewMediumText(string(yamlContent)),
		ZeropsYamlSetup:             types.NewStringNull(setup),
	})
	if err != nil {
		return err
	}
	if _, err = resp.Output(); err != nil {
		return errorsx.Convert(
			err,
			errorsx.And(
				errorsx.ErrorCode(errorsx.CodeZeropsYamlSetupNotFound),
				errorsx.Meta(func(_ errorsx.ApiError, metaItem map[string]interface{}) string {
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

// DeployAppVersion deploys an already-built app version (zcli deploy).
func DeployAppVersion(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	appVersionId entity.AppVersionId,
	zeropsYaml []byte,
	setup string,
) (entity.Process, error) {
	response, err := restApiClient.PutAppVersionDeploy(
		ctx,
		path.AppVersionId{Id: uuid.AppVersionId(appVersionId)},
		body.PutAppVersionDeploy{
			ZeropsYaml:      types.NewMediumTextNull(string(zeropsYaml)),
			ZeropsYamlSetup: types.NewStringNull(setup),
		},
	)
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return ProcessFromApiOutput(processOutput), nil
}

// BuildAndDeployAppVersion builds and deploys an app version (zcli push).
func BuildAndDeployAppVersion(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	appVersionId entity.AppVersionId,
	zeropsYaml []byte,
	setup string,
) (entity.Process, error) {
	response, err := restApiClient.PutAppVersionBuildAndDeploy(
		ctx,
		path.AppVersionId{Id: uuid.AppVersionId(appVersionId)},
		body.PutAppVersionBuildAndDeploy{
			ZeropsYaml:      types.MediumText(zeropsYaml),
			ZeropsYamlSetup: types.NewStringNull(setup),
		},
	)
	if err != nil {
		return entity.Process{}, err
	}
	processOutput, err := response.Output()
	if err != nil {
		return entity.Process{}, err
	}
	return ProcessFromApiOutput(processOutput), nil
}
