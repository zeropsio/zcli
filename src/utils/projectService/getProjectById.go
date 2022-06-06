package projectService

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func getById(ctx context.Context, _ business.ZeropsApiProtocolClient, projectId string) (*business.Project, error) {
	zdk := sdk.New(
		sdkBase.DefaultConfig(),
		http.DefaultClient,
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, "B7PUuq3LTZKo8c3gv1TUCwlS1SXUA7TyKl0gnyKxVFNQ")
	projectResponse, err := authorizedSdk.GetProject(ctx, path.ProjectId{Id: uuid.ProjectId(projectId)})
	if err != nil {
		return nil, err
	}

	project, err := projectResponse.Output()
	if err != nil {
		return nil, err
	}
	fmt.Println(project)

	bp := &business.Project{}
	err = RecastType(project, bp)
	if err != nil {
		log.Panic("Error recasting u1 to u2", err)
	}

	return bp, nil
}

// RecastType func converts types from one struct to another
func RecastType(a, b interface{}) error {
	js, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, b)
}
