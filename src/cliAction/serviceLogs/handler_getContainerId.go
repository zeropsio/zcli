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

func (h *Handler) getContainerId(ctx context.Context, sdkConfig sdkConfig.Config, clientId string, serviceId string, containerIndex int) (string, error) {
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
	})

	var sortData []body.EsSortItem
	sortData = append(sortData, body.EsSortItem{
		Name:      "number",
		Ascending: types.NewBoolNull(true),
	})

	response, err := authorizedSdk.PostContainerSearch(ctx, body.EsFilter{
		Search: searchData,
		Sort:   sortData,
	})
	if err != nil {
		return "", err
	}

	resOutput, err := response.Output()
	if err != nil { // TODO parse meta data
		return "", err
	}
	containers := resOutput.Items
	count := len(containers)

	if count == 0 {
		return "", fmt.Errorf("%s", i18n.LogNoContainerFound)
	}
	if count < containerIndex {
		verb, plural := "are", "s"
		if count < 2 {
			verb, plural = "is", ""
		}
		msg := fmt.Sprintf(i18n.LogTooFewContainers, verb, count, plural)
		return "", fmt.Errorf("%s", msg)
	}

	containerId := containers[containerIndex-1].Id

	return string(containerId), nil
}
