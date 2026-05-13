// Package zeropsRestApiClient provides a client for the zerops rest api
package zeropsRestApiClient

import (
	"net/http"
	"strings"

	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
)

type Handler struct {
	sdk.Handler
	env sdkBase.Environment
}

// NewAuthorizedClient builds an authorized SDK client. baseURL may be either a
// bare host ("api.example.com") or a full URL ("https://api.example.com",
// "http://127.0.0.1:1234"); a missing scheme defaults to https.
func NewAuthorizedClient(token string, baseURL string) *Handler {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	config := sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(baseURL))

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	return &Handler{
		Handler: sdk.AuthorizeSdk(sdk.New(config, httpClient), token),
		// temporary solution, I need my own endpoints
		env: sdkBase.NewEnvironment(config, httpClient).Authorize(token),
	}
}
