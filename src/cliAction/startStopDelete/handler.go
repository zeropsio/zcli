package startStopDelete

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/zipClient"
)

type Config struct {
}

type Method func(ctx context.Context, projectId string, serviceId string) (string, error)

type CmdType struct {
	Start   string
	Finish  string
	Execute Method
}

type RunConfig struct {
	ProjectName string
	ServiceName string
	Confirm     bool
	ParentCmd   constants.ParentCmd
	CmdData     CmdType
}

func (c *RunConfig) getCmdProps() (string, string, Method) {
	cd := c.CmdData
	return cd.Start, cd.Finish, cd.Execute
}

type Handler struct {
	config        Config
	httpClient    *httpClient.Handler
	zipClient     *zipClient.Handler
	apiGrpcClient business.ZeropsApiProtocolClient
}

func New(
	config Config,
	httpClient *httpClient.Handler,
	zipClient *zipClient.Handler,
	apiGrpcClient business.ZeropsApiProtocolClient,
) *Handler {
	return &Handler{
		config:        config,
		httpClient:    httpClient,
		zipClient:     zipClient,
		apiGrpcClient: apiGrpcClient,
	}
}
