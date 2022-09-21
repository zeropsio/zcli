package serviceLogs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) getServiceLogResData(ctx context.Context, sdkConfig sdkConfig.Config, projectId string) (string, string, error) {
	zdk := sdk.New(
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(sdkConfig.RegionUrl)),
		&http.Client{Timeout: 1 * time.Minute},
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, sdkConfig.Token)

	response, err := authorizedSdk.GetProjectLog(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return "", "", err
	}

	resOutput, err := response.Output()
	if err != nil {
		return "", "", fmt.Errorf("%s %v", i18n.LogAccessFailed, err)
	}
	method, url := getLogRequestData(resOutput)
	return method, url, nil
}
