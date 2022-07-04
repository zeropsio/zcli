package importProjectService

import (
	"context"

	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) getClientId(ctx context.Context, config RunConfig) (string, error) {

	if len(config.ClientId) > 0 {
		return config.ClientId, nil
	}

	return projectService.GetClientId(ctx, h.apiGrpcClient)
}
