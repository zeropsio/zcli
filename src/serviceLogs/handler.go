package serviceLogs

import (
	"context"
	"io"
	"os"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/types/uuid"
)

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
	Tags           []string
}

type Handler struct {
	out           io.Writer
	restApiClient *zeropsRestApiClient.Handler

	lastMsgId string
}

func (h *Handler) Writer() io.Writer {
	return h.out
}

func NewStdout(restApiClient *zeropsRestApiClient.Handler) *Handler {
	return New(os.Stdout, restApiClient)
}

func New(out io.Writer, restApiClient *zeropsRestApiClient.Handler) *Handler {
	return &Handler{
		restApiClient: restApiClient,
		out:           out,
	}
}

func (h *Handler) Run(ctx context.Context, config RunConfig) error {
	inputs, err := h.checkInputValues(config)
	if err != nil {
		return err
	}
	return h.printLogs(ctx, inputs, config.Project.Id, config.ServiceId, config.Container.Id)
}
