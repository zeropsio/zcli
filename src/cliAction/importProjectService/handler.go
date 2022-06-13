package importProjectService

import (
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
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
	sdkConfig     sdkConfig.Config
}

func New(config Config, httpClient *httpClient.Handler, apiGrpcClient business.ZeropsApiProtocolClient, sdkConfig sdkConfig.Config) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		apiGrpcClient: apiGrpcClient,
		sdkConfig:     sdkConfig,
	}
}
