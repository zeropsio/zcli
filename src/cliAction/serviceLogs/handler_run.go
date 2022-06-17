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
	// TODO 1. check inout values
	fmt.Printf("%v \n", config)
	// TODO 2. implement logic

	return nil
}
