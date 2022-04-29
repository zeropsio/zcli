package startStopDeleteProject

import "context"

func (h *Handler) Run(ctx context.Context, config RunConfig, actionType string) error {
	projectId, err := h.getProjectId(ctx, config)
	if err != nil {
		return err
	}

	if actionType == "start" {
		return h.RunStart(ctx, config, projectId)
	} else if actionType == "stop" {
		return h.RunStop(ctx, config, projectId)
	} else {
		return h.RunDelete(ctx, config, projectId, actionType)
	}

}
