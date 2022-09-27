package projectService

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func getById(ctx context.Context, sdkConfig sdkConfig.Config, projectId string) (string, error) {
	zdk := sdk.New(
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(sdkConfig.RegionUrl)),
		&http.Client{Timeout: 1 * time.Minute},
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, sdkConfig.Token)
	projectResponse, err := authorizedSdk.GetProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return "", err
	}

	project, err := projectResponse.Output()
	if err != nil {
		if apiError.HasErrorCode(err, errorCode.ProjectNotFound) {
			return "", fmt.Errorf("%s. %s", i18n.ProjectNotFound, i18n.ProjectWrongId)
		}
		return "", err
	}

	return string(project.Id), nil
}
