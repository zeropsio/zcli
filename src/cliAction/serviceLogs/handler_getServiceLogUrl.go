package serviceLogs

import (
	"context"
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
	"net/http"
	"time"
)

func (h *Handler) getServiceLogData(ctx context.Context, sdkConfig sdkConfig.Config, projectId string) (string, string, types.DateTime, error) {
	zdk := sdk.New(
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(sdkConfig.RegionUrl)),
		&http.Client{Timeout: 1 * time.Minute},
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, sdkConfig.Token)

	response, err := authorizedSdk.GetProjectLog(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return "", "", types.DateTime{}, err
	}

	resOutput, err := response.Output()
	if err != nil { // TODO parse meta data
		return "", "", types.DateTime{}, fmt.Errorf("%s %v", i18n.LogAccessFailed, err)
	}
	method, url, expiration := getLogRequestData(resOutput)
	return method, url, expiration, nil
}
