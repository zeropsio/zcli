package serviceLogs

import (
	"context"
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
	if err != nil {
		return err
	}

	limit, minSeverity, facility, format, formatTemplate, err := h.checkInputValues(config)
	if err != nil {
		return err
	}
	serviceName, source, containerIndex, err := h.getNameSourceContainerId(config)
	if err != nil {
		return err
	}

	mode := RESPONSE
	if config.Follow {
		mode = STREAM
	}
	fmt.Println(limit, minSeverity, facility, format, formatTemplate, serviceName, source, containerIndex, mode)

	service, err := projectService.GetServiceStack(ctx, h.apiGrpcClient, projectId, serviceName)
	if err != nil {
		return err
	}

	serviceTypeCategory := service.GetServiceStackTypeInfo().GetServiceStackTypeCategory().String()
	fmt.Println("service category", serviceTypeCategory)
	if serviceTypeCategory != USER {
		return fmt.Errorf("%s", i18n.LogRuntimeOnly)
	}
	serviceId := service.GetId()
	if containerIndex > 0 {
		fmt.Println(containerIndex)
		containerId, err := h.getContainerId(ctx, h.sdkConfig, serviceId, containerIndex)
		if err != nil {
			return err
		}
		fmt.Println(containerId)
	}

	logServiceId := serviceId
	fmt.Println("source", source)
	if source == BUILD {
		logServiceId, err = h.getAppVersionServiceId(ctx, h.sdkConfig, serviceId)
		if err != nil {
			return err
		}
	}
	fmt.Println("log service id ", logServiceId)
	// TODO get logs

	return nil
}
