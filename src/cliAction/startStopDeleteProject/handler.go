package startStopDeleteProject

import (
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"
)

type Config struct {
}

type RunConfig struct {
	ProjectName string
	ServiceName string
	Confirm     bool
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	zipClient     *zipClient.Handler
	apiGrpcClient business.ZeropsApiProtocolClient
}

func New(
	config Config,
	httpClient *httpClient.Handler,
	zipClient *zipClient.Handler,
	apiGrpcClient business.ZeropsApiProtocolClient,
) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		zipClient:     zipClient,
		apiGrpcClient: apiGrpcClient,
	}
}
