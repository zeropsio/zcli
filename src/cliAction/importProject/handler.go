package importProject

import (
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

const importYamlFileName = "import.yml"

type Config struct {
}

type RunConfig struct {
	WorkingDir     string
	ZipFilePath    string
	VersionName    string
	ImportYamlPath *string
	ClientId       string
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
