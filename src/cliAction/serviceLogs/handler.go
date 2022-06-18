package serviceLogs

import (
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
)

type Config struct {
}

type Levels map[string]string

type RunConfig struct {
	ProjectNameOrId string
	ServiceName     string
	Limit           uint32
	MinSeverity     string
	MsgType         string
	Format          string
	FormatTemplate  string
	Follow          bool
	Levels          Levels
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
