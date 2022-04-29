package startStopDeleteProject

import "context"

func (h *Handler) Run(ctx context.Context, config RunConfig, actionType string) error {
	projectId, err := h.getProjectId(ctx, config)
	if err != nil {
		return err
	}

	// todo add delete and change to switch
	if actionType == "start" {
		return h.RunStart(ctx, config, projectId)
	} else {
		return h.RunStop(ctx, config, projectId)
	}

}
