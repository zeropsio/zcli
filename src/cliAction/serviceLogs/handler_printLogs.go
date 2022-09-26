package serviceLogs

import (
	"context"
	"fmt"
	"strings"
)

func (h *Handler) printLogs(ctx context.Context, config RunConfig, inputs InputValues, containerId, logServiceId, projectId string) error {
	method, url, err := h.getServiceLogResData(ctx, h.sdkConfig, projectId)
	if err != nil {
		return err
	}

	query := makeQueryParams(inputs, logServiceId, containerId)

	if inputs.mode == RESPONSE {
		err = getLogs(ctx, method, HTTPS+url+query, inputs.format, inputs.formatTemplate, inputs.mode)
		if err != nil {
			return err
		}
	}
	if inputs.mode == STREAM {
		wsUrl := getWsUrl(url)
		err := h.getLogStream(ctx, config, inputs, wsUrl, query, containerId, logServiceId, projectId)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeQueryParams(inputs InputValues, logServiceId, containerId string) string {
	query := fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s",
		inputs.limit, getDesc(inputs.mode), inputs.facility, logServiceId)

	if inputs.minSeverity != -1 {
		query += fmt.Sprintf("&minimumSeverity=%d", inputs.minSeverity)
	}

	if containerId != "" {
		query += fmt.Sprintf("&containerId=%s", containerId)
	}

	return query
}

func getDesc(mode string) int {
	if mode == RESPONSE {
		return 1
	}
	return 0
}

func getWsUrl(apiUrl string) string {
	urlSplit := strings.Split(apiUrl, "?")
	url, token := urlSplit[0], urlSplit[1]
	return url + "/stream?" + token
}
