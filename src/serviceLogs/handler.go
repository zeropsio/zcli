package serviceLogs

import (
	"io"
	"os"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type Config struct {
}

type Levels [8][2]string

var DefaultLevels = Levels{
	{"EMERGENCY", "0"},
	{"ALERT", "1"},
	{"CRITICAL", "2"},
	{"ERROR", "3"},
	{"WARNING", "4"},
	{"NOTICE", "5"},
	{"INFORMATIONAL", "6"},
	{"DEBUG", "7"},
}

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
	Tags           []string
}

type Handler struct {
	out           io.Writer
	config        Config
	restApiClient *zeropsRestApiClient.Handler

	lastMsgId string
}

func (h *Handler) Writer() io.Writer {
	return h.out
}

func NewStdout(config Config, restApiClient *zeropsRestApiClient.Handler) *Handler {
	return New(os.Stdout, config, restApiClient)
}

func New(out io.Writer, config Config, restApiClient *zeropsRestApiClient.Handler) *Handler {
	return &Handler{
		config:        config,
		restApiClient: restApiClient,
		out:           out,
	}
}
