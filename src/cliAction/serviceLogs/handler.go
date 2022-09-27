package serviceLogs

import (
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
)

type Config struct {
}

type Levels [8][2]string

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
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient
	sdkConfig     sdkConfig.Config
	LastMsgId     string
}

func New(config Config, httpClient *httpClient.Handler, apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient, sdkConfig sdkConfig.Config) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		apiGrpcClient: apiGrpcClient,
		sdkConfig:     sdkConfig,
	}
}
