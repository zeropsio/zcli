// Package zeropsRestApiClient provides a client for the zerops rest api
package zeropsRestApiClient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/stringId"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h *Handler) GetServiceStackByProject(
	ctx context.Context,
	projectId uuid.ProjectId,
	clientId uuid.ClientId,
) (GetServiceStackByProjectResponse, error) {
	var response GetServiceStackByProjectResponse

	u := "/api/rest/public/service-stack/search"

	filter := EsFilter{
		Search: []EsSearchItem{
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    projectId.Native(),
			},
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

type GetServiceStackByProjectResponse struct {
	success            EsServiceStackResponse
	err                error
	responseHeaders    http.Header
	responseStatusCode int
}

func (r GetServiceStackByProjectResponse) Output() (output EsServiceStackResponse, err error) {
	return r.success, r.err
}

type EsServiceStackResponse struct {
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
	TotalHits int              `json:"totalHits"`
	Items     []EsServiceStack `json:"items"`
}

type EsServiceStack struct {
	Id                        uuid.ServiceStackId                `json:"id"`
	ProjectId                 uuid.ProjectId                     `json:"projectId"`
	ClientId                  uuid.ClientId                      `json:"clientId"`
	ServiceStackTypeId        stringId.ServiceStackTypeId        `json:"serviceStackTypeId"`
	ServiceStackTypeVersionId stringId.ServiceStackTypeVersionId `json:"serviceStackTypeVersionId"`
	Status                    enum.ServiceStackStatusEnum        `json:"status"`
	Name                      types.String                       `json:"name"`
	IsSystem                  types.Bool                         `json:"isSystem"`
	ServiceStackTypeInfo      EsServiceStackInfoJsonObject       `json:"serviceStackTypeInfo"`
}

type EsServiceStackInfoJsonObject struct {
	ServiceStackTypeName        types.String                      `json:"serviceStackTypeName"`        // serviceStackTypeName - types.String
	ServiceStackTypeCategory    enum.ServiceStackTypeCategoryEnum `json:"serviceStackTypeCategory"`    // serviceStackTypeCategory - enum.ServiceStackTypeCategoryEnum
	ServiceStackTypeVersionName types.String                      `json:"serviceStackTypeVersionName"` // serviceStackTypeVersionName - types.String

}
