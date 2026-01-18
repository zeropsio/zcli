package serviceLogs

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) getServiceLogResData(ctx context.Context, projectId uuid.ProjectId) (string, string, error) {
	response, err := h.restApiClient.GetProjectLog(ctx, path.ProjectId{Id: projectId}, query.GetProjectLog{})
	if err != nil {
		return "", "", err
	}

	resOutput, err := response.Output()
	if err != nil {
		return "", "", errors.Errorf("%s %v", i18n.T(i18n.LogAccessFailed), err)
	}
	return http.MethodGet, resOutput.Url.String(), nil
}
