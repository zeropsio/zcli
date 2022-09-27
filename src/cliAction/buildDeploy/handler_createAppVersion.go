package buildDeploy

import (
	"context"

	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
)

func (h *Handler) createAppVersion(ctx context.Context, config RunConfig, serviceStack *zBusinessZeropsApiProtocol.GetServiceStackByNameResponseDto) (*zBusinessZeropsApiProtocol.PostAppVersionResponseDto, error) {
	appVersionResponse, err := h.apiGrpcClient.PostAppVersion(ctx, &zBusinessZeropsApiProtocol.PostAppVersionRequest{
		ServiceStackId: serviceStack.GetId(),
		Name: func() *zBusinessZeropsApiProtocol.StringNull {
			if config.VersionName != "" {
				return &zBusinessZeropsApiProtocol.StringNull{
					Value: config.VersionName,
					Valid: true,
				}
			}
			return &zBusinessZeropsApiProtocol.StringNull{}
		}(),
	})
	if err := proto.BusinessError(appVersionResponse, err); err != nil {
		return nil, err
	}
	appVersion := appVersionResponse.GetOutput()

	return appVersion, nil
}
