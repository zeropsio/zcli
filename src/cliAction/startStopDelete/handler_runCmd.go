package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/utils/processChecker"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectName)
	if err != nil {
		return err
	}
	var serviceId string

	if config.ParentCmd == constants.Service {
		serviceId, err = projectService.GetServiceId(ctx, h.apiGrpcClient, projectId, config.ServiceName)
		if err != nil {
			return err
		}
	}

	err = h.runCmd(ctx, config, projectId, serviceId)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) runCmd(ctx context.Context, config RunConfig, projectId string, serviceId string) error {

	startMsg, finishMsg, callback := h.getCmdType(config.ParentCmd, config.ChildCmd)
	msg := GetConfirm(config)
	if len(msg) > 0 {
		fmt.Println(msg)
		return nil
	}
	fmt.Println(startMsg)

	processId, err := callback(ctx, h, projectId, serviceId)
	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(constants.Success + finishMsg)

	return nil
}
