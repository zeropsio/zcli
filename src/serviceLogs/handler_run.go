package serviceLogs

import (
	"context"

	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return err
	}

	// TODO - janhajek check empty containerID
	if err = h.printLogs(
		ctx,
		inputs,
		uuid.ProjectId(config.Project.Id),
		uuid.ServiceStackId(config.ServiceId),
		uuid.ContainerId(config.Container.Id),
	); err != nil {
		return err
	}

	return nil
}
