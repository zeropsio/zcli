package buildDeploy

import (
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

type Config struct {
}

type RunConfig struct {
	ProjectName      string
	ServiceStackName string
	PathsForPacking  []string
	WorkingDir       string
	ZipFilePath      string
	VersionName      string
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	zipClient     *zipClient.Handler
	apiGrpcClient zeropsApiProtocol.ZeropsApiProtocolClient
}

func New(
	config Config,
	httpClient *httpClient.Handler,
	zipClient *zipClient.Handler,
	apiGrpcClient zeropsApiProtocol.ZeropsApiProtocolClient,
) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		zipClient:     zipClient,
		apiGrpcClient: apiGrpcClient,
	}
}
