// Package zeropsRestApiClient provides a client for the zerops rest api
package zeropsRestApiClient

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) GetProjectsByClient(
	ctx context.Context,
	clientId uuid.ClientId,
) (GetProjectsByClientResponse, error) {
	var response GetProjectsByClientResponse

	u := "/api/rest/public/project/search"

	filter := EsFilter{
		Search: []EsSearchItem{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    clientId.Native(),
			},
		},
	}

	sdkResponse := sdkBase.Post(ctx, h.env, u, filter)

	if sdkResponse.Err != nil {
		return response, sdkResponse.Err
	}
	response.responseHeaders = sdkResponse.HttpResponse.Header
	response.responseStatusCode = sdkResponse.HttpResponse.StatusCode

	decoder := json.NewDecoder(sdkResponse.ResponseData)
	if sdkResponse.HttpResponse.StatusCode < http.StatusMultipleChoices {
		if err := decoder.Decode(&response.success); err != nil {
			return response, err
		}
	} else {
		responseString := sdkResponse.ResponseData.String()
		apiErrorResponse := struct {
			Error apiError.Error `json:"error"`
		}{}
		err := decoder.Decode(&apiErrorResponse)
		if err != nil {
			return response, errors.New(sdkResponse.HttpResponse.Status + ": " + responseString)
		}
		apiErrorResponse.Error.HttpStatusCode = sdkResponse.HttpResponse.StatusCode
		response.err = apiErrorResponse.Error
	}

	return response, nil
}

type EsFilter struct {
	Search []EsSearchItem `json:"search"`
}

type EsSearchItem struct {
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type GetProjectsByClientResponse struct {
	success            EsProjectResponse
	err                error
	responseHeaders    http.Header
	responseStatusCode int
}

func (r GetProjectsByClientResponse) Output() (output EsProjectResponse, err error) {
	return r.success, r.err
}

type EsProjectResponse struct {
	Limit     int64       `json:"limit"`
	Offset    int64       `json:"offset"`
	TotalHits int64       `json:"totalHits"`
	Items     []EsProject `json:"items"`
}

type EsProject struct {
	Id          uuid.ProjectId         `json:"id"`
	ClientId    uuid.ClientId          `json:"clientId"`
	Name        types.String           `json:"name"`
	Description types.TextNull         `json:"description"`
	Status      enum.ProjectStatusEnum `json:"status"`
}
