package serviceLogs

import (
	"context"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return err
	}

	// TODO - janhajek check empty containerID
	if err = h.printLogs(ctx, inputs, config.Project.Id, config.ServiceId, config.Container.Id); err != nil {
		return err
	}

	return nil
}
