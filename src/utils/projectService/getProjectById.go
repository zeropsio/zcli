package projectService

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeropsio/zerops-go/apiError"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func getById(ctx context.Context, sdkConfig sdkConfig.Config, projectId string) (string, error) {
	zdk := sdk.New(
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(sdkConfig.RegionUrl)),
		http.DefaultClient,
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, sdkConfig.Token)
	projectResponse, err := authorizedSdk.GetProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return "", err
	}

	project, err := projectResponse.Output()
	if err != nil { // TODO try to parse meta data
		var apiErr apiError.Error
		if errors.As(err, &apiErr) {
			fmt.Println("apiError detected") //FIXME
		}
		if strings.Contains(err.Error(), "Invalid user input") {
			return "", fmt.Errorf("%s", i18n.ProjectIdInvalid)
		}
		if strings.Contains(err.Error(), "Project not found") {
			return "", fmt.Errorf("%s. %s", i18n.ProjectNotFound, i18n.ProjectWrongId)
		}
		return "", err
	}

	return string(project.Id), nil
}
