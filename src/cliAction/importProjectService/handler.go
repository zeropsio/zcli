package importProjectService

import (
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
)

type Config struct {
}

type RunConfig struct {
	WorkingDir      string
	ImportYamlPath  string
	ClientId        string
	ProjectNameOrId string
	ParentCmd       constants.ParentCmd
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	apiGrpcClient business.ZeropsApiProtocolClient
}

func New(config Config, httpClient *httpClient.Handler, apiGrpcClient business.ZeropsApiProtocolClient) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		apiGrpcClient: apiGrpcClient,
	}
}
