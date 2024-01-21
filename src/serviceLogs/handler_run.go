package serviceLogs

import (
	"context"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return err
	}

	// FIXME - janhajek check empty containerID
	if err = h.printLogs(ctx, inputs, config.Project.ID, config.ServiceId, config.Container.ID); err != nil {
		return err
	}

	return nil
}
