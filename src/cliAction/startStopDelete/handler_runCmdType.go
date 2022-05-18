package startStopDelete

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) Run(ctx context.Context, config RunConfig, parentCmd constants.ParentCmd, actionType string) error {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectName)
	if err != nil {
		return err
	}

	if parentCmd == constants.Project {
		if actionType == constants.Start {
			return h.ProjectStart(ctx, projectId)
		} else if actionType == constants.Stop {
			return h.ProjectStop(ctx, projectId)
		} else {
			return h.ProjectDelete(ctx, projectId, config)
		}
	}

	serviceId, err := projectService.GetServiceId(ctx, h.apiGrpcClient, projectId, config.ServiceName)
	if err != nil {
		return err
	}

	if actionType == constants.Start {
		return h.ServiceStart(ctx, serviceId)
	} else if actionType == constants.Stop {
		return h.ServiceStop(ctx, serviceId)
	} else {
		return h.ServiceDelete(ctx, serviceId, config)
	}
}
