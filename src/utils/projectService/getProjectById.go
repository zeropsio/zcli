package projectService

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func getById(ctx context.Context, token, projectId string) (string, error) {
	zdk := sdk.New(
		// FIXME remove custom url for production
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint("https://demo.devel.zerops.io")),
		http.DefaultClient,
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, token)
	projectResponse, err := authorizedSdk.GetProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return "", err
	}

	project, err := projectResponse.Output()
	if err != nil { // FIXME try to parse meta data
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
