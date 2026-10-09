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
}

// NewClient builds an SDK client without authorization, for endpoints like the device login or the token refresh.
// baseURL may be either a bare host ("api.example.com") or a full URL ("https://api.example.com", "http://127.0.0.1:1234"), a missing scheme defaults to https.
func NewClient(baseURL string) *Handler {
	return newHandler(baseURL, newTransport())
}

// NewAuthorizedClient builds an SDK client that sends the given static token.
func NewAuthorizedClient(token string, baseURL string) *Handler {
	return NewTokenSourceClient(StaticToken(token), baseURL)
}

// NewTokenSourceClient builds an SDK client that takes its token from source and retries a refused request once with a renewed token.
func NewTokenSourceClient(source TokenSource, baseURL string) *Handler {
	return newHandler(baseURL, &authTransport{base: newTransport(), source: source})
}

func newHandler(baseURL string, transport http.RoundTripper) *Handler {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	config := sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(baseURL))
	return &Handler{
		Handler: sdk.New(config, &http.Client{Transport: transport}),
	}
}

func newTransport() http.RoundTripper {
	return &http.Transport{
		DisableKeepAlives: true,
	}
}
