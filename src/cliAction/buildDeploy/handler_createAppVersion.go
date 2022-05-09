package buildDeploy

import (
	"context"

	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) createAppVersion(ctx context.Context, config RunConfig, serviceStack *business.GetServiceStackByNameResponseDto) (*business.PostAppVersionResponseDto, error) {
	appVersionResponse, err := h.apiGrpcClient.PostAppVersion(ctx, &business.PostAppVersionRequest{
		ServiceStackId: serviceStack.GetId(),
		Name: func() *business.StringNull {
			if config.VersionName != "" {
				return &business.StringNull{
					Value: config.VersionName,
					Valid: true,
				}
			}
			return &business.StringNull{}
		}(),
	})
	if err := proto.BusinessError(appVersionResponse, err); err != nil {
		return nil, err
	}
	appVersion := appVersionResponse.GetOutput()

	return appVersion, nil
}
