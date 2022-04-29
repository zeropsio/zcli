package startStopDeleteProject

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
)

func (h *Handler) Run(ctx context.Context, config RunConfig, actionType string) error {
	projectId, err := h.getProjectId(ctx, config)
	if err != nil {
		return err
	}

	if actionType == constants.Start {
		return h.RunStart(ctx, config, projectId)
	} else if actionType == constants.Stop {
		return h.RunStop(ctx, config, projectId)
	} else {
		return h.RunDelete(ctx, config, projectId, actionType)
	}

}
