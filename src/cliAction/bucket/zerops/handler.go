package bucketZerops

import (
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/sdkConfig"
)

type Config struct {
}

type RunConfig struct {
	ProjectNameOrId  string
	ServiceStackName string
	BucketName       string
	XAmzAcl          string
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient
	sdkConfig     sdkConfig.Config
}

func New(
	config Config,
	httpClient *httpClient.Handler,
	apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient,
	sdkConfig sdkConfig.Config,
) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		apiGrpcClient: apiGrpcClient,
		sdkConfig:     sdkConfig,
	}
}
