package serviceLogs

import (
	"context"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return NewInvalidRequestError("Run", "invalid input values", err)
	}

	if config.Container.ID == "" {
		return NewInvalidRequestError("Run", "container ID is required", nil)
	}

	if err = h.printLogs(ctx, inputs, config.Project.ID, config.ServiceId, config.Container.ID); err != nil {
		if IsInvalidRequestError(err) {
			return err
		}
		if IsLogResponseError(err) {
			return err
		}
		return NewLogResponseError(0, "failed to print logs", err)
	}

	return nil
}
