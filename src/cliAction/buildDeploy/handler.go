package buildDeploy

import (
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/archiveClient"
	"github.com/zeropsio/zcli/src/utils/httpClient"
	"github.com/zeropsio/zcli/src/utils/sdkConfig"
)

const zeropsYamlFileName = "zerops.yml"

type Config struct {
}

type RunConfig struct {
	ProjectNameOrId  string
	SourceName       string
	ServiceStackName string
	PathsForPacking  []string
	WorkingDir       string
	ArchiveFilePath  string
	VersionName      string
	ZeropsYamlPath   string
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	archClient    *archiveClient.Handler
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient
	sdkConfig     sdkConfig.Config
}

func New(
	config Config,
	httpClient *httpClient.Handler,
	archClient *archiveClient.Handler,
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient,
	sdkConfig sdkConfig.Config,
) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		archClient:    archClient,
		apiGrpcClient: apiGrpcClient,
		sdkConfig:     sdkConfig,
	}
}
