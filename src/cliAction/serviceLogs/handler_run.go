package serviceLogs

import (
	"context"
	"fmt"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/projectService"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return err
	}

	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
	if err != nil {
		return err
	}

	serviceName, source, containerIndex, err := h.getNameSourceContainerId(config)
	if err != nil {
		return err
	}

	service, err := projectService.GetServiceStack(ctx, h.apiGrpcClient, projectId, serviceName)
	if err != nil {
		return err
	}

	serviceTypeCategory := service.GetServiceStackTypeInfo().GetServiceStackTypeCategory().String()

	if serviceTypeCategory != USER {
		return fmt.Errorf("%s", i18n.LogRuntimeOnly)
	}
	serviceId := service.GetId()
	containerId := ""
	// defined by user, can be 1 or higher
	if containerIndex > 0 {
		containerId, err = h.getContainerId(ctx, h.sdkConfig, serviceId, containerIndex)
		if err != nil {
			return err
		}
	}

	logServiceId := serviceId
	if source == BUILD {
		logServiceId, err = h.getAppVersionServiceId(ctx, h.sdkConfig, serviceId)
		if err != nil {
			return err
		}
	}

	method, url, expiration, err := h.getServiceLogResData(ctx, h.sdkConfig, projectId)
	if err != nil {
		return err
	}

	query := makeQueryParams(inputs, logServiceId, containerId)

	if inputs.mode == RESPONSE {
		err = getLogs(ctx, method, HTTP+url+query, inputs.format, inputs.formatTemplate, inputs.mode)
		if err != nil {
			return err
		}
	}
	if inputs.mode == STREAM {
		// TODO extract to a func prepareStreamUrl
		urlSplit := strings.Split(url, "?")
		url = urlSplit[0]
		token := urlSplit[1]
		url = url + "/stream?" + token
		fmt.Println("url fixed is", url)
		err := getLogStream(ctx, expiration, inputs.format, url, query, inputs.mode)
		if err != nil {
			return err
		}
	}

	return nil
}
