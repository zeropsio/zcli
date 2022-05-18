package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
)

func (h *Handler) Run(ctx context.Context, config RunConfig, parentCmd string, actionType string) error {
	projectId, err := h.getProjectId(ctx, config)
	if err != nil {
		fmt.Println("error ", err)
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

	serviceId, err := h.getServiceId(ctx, config, projectId)
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
