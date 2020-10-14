package buildDeploy

import (
	"context"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) createAppVersion(ctx context.Context, config RunConfig, serviceStack *zeropsApiProtocol.GetServiceStackByNameResponseDto) (*zeropsApiProtocol.PostAppVersionResponseDto, error) {
	appVersionResponse, err := h.apiGrpcClient.PostAppVersion(ctx, &zeropsApiProtocol.PostAppVersionRequest{
		ServiceStackId: serviceStack.GetId(),
		Name: func() *zeropsApiProtocol.StringNull {
			if config.VersionName != "" {
				return &zeropsApiProtocol.StringNull{
					Value: config.VersionName,
					Valid: true,
				}
			}
			return &zeropsApiProtocol.StringNull{}
		}(),
	})
	if err := utils.HandleGrpcApiError(appVersionResponse, err); err != nil {
		return nil, err
	}
	appVersion := appVersionResponse.GetOutput()

	return appVersion, nil
}
