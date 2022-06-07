package buildDeploy

import (
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/archiveClient"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
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
	ZeropsYamlPath   *string
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	archClient    *archiveClient.Handler
	apiGrpcClient business.ZeropsApiProtocolClient
	sdkConfig     sdkConfig.Config
}

func New(
	config Config,
	httpClient *httpClient.Handler,
	archClient *archiveClient.Handler,
	apiGrpcClient business.ZeropsApiProtocolClient,
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
