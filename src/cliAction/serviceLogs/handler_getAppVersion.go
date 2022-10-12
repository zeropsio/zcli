package serviceLogs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types"
)

func (h *Handler) getAppVersionServiceId(ctx context.Context, sdkConfig sdkConfig.Config, clientId string, serviceId string) (string, error) {
	zdk := sdk.New(
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(sdkConfig.RegionUrl)),
		&http.Client{Timeout: 1 * time.Minute},
	)

	authorizedSdk := sdk.AuthorizeSdk(zdk, sdkConfig.Token)
	var searchData []body.EsSearchItem
	searchData = append(searchData, body.EsSearchItem{
		Name:     "clientId",
		Operator: "eq",
		Value:    types.String(clientId),
	}, body.EsSearchItem{
		Name:     "serviceStackId",
		Operator: "eq",
		Value:    types.String(serviceId),
	}, body.EsSearchItem{
		Name:     "build.serviceStackId",
		Operator: "ne",
		Value:    "",
	})

	var sortData []body.EsSortItem
	sortData = append(sortData, body.EsSortItem{
		Name:      "sequence",
		Ascending: types.NewBoolNull(false),
	})

	response, err := authorizedSdk.PostAppVersionSearch(ctx, body.EsFilter{
		Search: searchData,
		Sort:   sortData,
		Limit:  types.NewIntNull(1),
	})
	if err != nil {
		return "", err
	}

	resOutput, err := response.Output()
	if err != nil {
		return "", err
	}

	if len(resOutput.Items) == 0 {
		return "", fmt.Errorf("%s", i18n.LogNoBuildFound)
	}

	app := resOutput.Items[0]
	status := app.Status
	if status == UPLOADING || app.Build == nil {
		return "", fmt.Errorf("%s", i18n.LogBuildStatusUploading)
	}

	id, _ := app.Build.ServiceStackId.Get()

	return string(id), nil
}
