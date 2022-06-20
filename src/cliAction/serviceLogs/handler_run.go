package serviceLogs

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
	if err != nil {
		return err
	}
	var serviceId string

	serviceId, err = projectService.GetServiceId(ctx, h.apiGrpcClient, projectId, config.ServiceName)
	if err != nil {
		return err
	}

	err = h.runCmd(ctx, config, serviceId)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) runCmd(_ context.Context, config RunConfig, serviceId string) error {
	fmt.Printf("service ID %v \n", serviceId)
	// TODO 1. check input values
	limit, err := h.getLimit(config)
	if err != nil {
		return err
	}
	fmt.Println("limit", limit)
	sev, err := h.getMinSeverity(config)
	if err != nil {
		return err
	}
	fmt.Println("severity", sev)

	facility, err := h.getFacility(config)
	if err != nil {
		return err
	}
	fmt.Println("facility", facility)
	format, formatTemplate, err := h.getFormat(config)
	if err != nil {
		return err
	}
	fmt.Println("template", format, formatTemplate)

	return nil
}
