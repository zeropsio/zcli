package serviceLogs

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeropsio/zerops-go/dto/output"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) getServiceLogResData(ctx context.Context, projectId uuid.ProjectId) (string, string, error) {
	response, err := h.restApiClient.GetProjectLog(ctx, path.ProjectId{Id: projectId})
	if err != nil {
		return "", "", err
	}

	resOutput, err := response.Output()
	if err != nil {
		return "", "", fmt.Errorf("%s %v", i18n.T(i18n.LogAccessFailed), err)
	}
	method, url := getLogRequestData(resOutput)
	return method, url, nil
}

func getLogRequestData(resOutput output.ProjectLog) (string, string) {
	outputUrl := string(resOutput.Url)
	urlData := strings.Split(outputUrl, " ")
	method, url := urlData[0], urlData[1]

	return method, url
}
