package startStopDeleteProject

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
)

func (h *Handler) Run(ctx context.Context, config RunConfig, actionType string) error {
	projectId, err := h.getProjectId(ctx, config)
	if err != nil {
		fmt.Println("error ", err)
		return err
	}

	if actionType == constants.Start {
		return h.RunStart(ctx, projectId)
	} else if actionType == constants.Stop {
		return h.RunStop(ctx, projectId)
	} else {
		return h.RunDelete(ctx, config, projectId)
	}

}
