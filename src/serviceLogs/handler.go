package serviceLogs

import (
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Config struct {
}

type Levels [8][2]string

type RunConfig struct {
	Project        entity.Project
	ServiceId      uuid.ServiceStackId
	Container      entity.Container
	Limit          int
	MinSeverity    string
	MsgType        string
	Format         string
	FormatTemplate string
	Follow         bool
	Levels         Levels
}

type Handler struct {
	config        Config
	restApiClient *zeropsRestApiClient.Handler

	lastMsgId string
}

func New(config Config, restApiClient *zeropsRestApiClient.Handler) *Handler {
	return &Handler{
		config:        config,
		restApiClient: restApiClient,
	}
}
