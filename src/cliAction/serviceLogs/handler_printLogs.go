package serviceLogs

import (
	"context"
	"strings"
)

func (h *Handler) printLogs(ctx context.Context, inputs InputValues, containerId, logServiceId, projectId string) error {
	method, url, err := h.getServiceLogResData(ctx, h.sdkConfig, projectId)
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
		wsUrl := getWsUrl(url)
		err := h.getLogStream(ctx, inputs, wsUrl, query, containerId, logServiceId, projectId)
		if err != nil {
			return err
		}
	}
	return nil
}

func getWsUrl(apiUrl string) string {
	urlSplit := strings.Split(apiUrl, "?")
	url, token := urlSplit[0], urlSplit[1]
	return url + "/stream?" + token
}
