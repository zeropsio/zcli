// Package zeropsRestApiClient provides a client for the zerops rest api
package zeropsRestApiClient

import (
	"net/http"

	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
)

type Handler struct {
	sdk.Handler
	env sdkBase.Environment
}

func NewAuthorizedClient(token string, regionUrl string) *Handler {
	config := sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(regionUrl))

	return &Handler{
		Handler: sdk.AuthorizeSdk(sdk.New(config, http.DefaultClient), token),
		// temporary solution, I need my own endpoints
		env: sdkBase.NewEnvironment(config, http.DefaultClient).Authorize(token),
	}
}
