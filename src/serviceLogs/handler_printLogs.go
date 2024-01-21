package serviceLogs

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) printLogs(
	ctx context.Context,
	inputs InputValues,
	projectId uuid.ProjectId,
	serviceId uuid.ServiceStackId,
	containerId uuid.ContainerId,
) error {
	method, url, err := h.getServiceLogResData(ctx, projectId)
	if err != nil {
		return err
	}

	query := makeQueryParams(inputs, serviceId, containerId)

	if inputs.mode == RESPONSE {
		err = getLogs(ctx, method, HTTPS+url+query, inputs.format, inputs.formatTemplate, inputs.mode)
		if err != nil {
			return err
		}
	}
	if inputs.mode == STREAM {
		wsUrl := getWsUrl(url)
		err := h.getLogStream(ctx, inputs, projectId, serviceId, containerId, wsUrl, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeQueryParams(inputs InputValues, serviceId uuid.ServiceStackId, containerId uuid.ContainerId) string {
	query := fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s",
		inputs.limit, getDesc(inputs.mode), inputs.facility, serviceId)

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
