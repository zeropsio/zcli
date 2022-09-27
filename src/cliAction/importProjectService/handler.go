package importProjectService

import (
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
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
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient
	sdkConfig     sdkConfig.Config
}

func New(config Config, httpClient *httpClient.Handler, apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient, sdkConfig sdkConfig.Config) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		apiGrpcClient: apiGrpcClient,
		sdkConfig:     sdkConfig,
	}
}
