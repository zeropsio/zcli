package stopVpn

import (
	"context"

	"github.com/zerops-io/zcli/src/service/logger"
	"github.com/zerops-io/zcli/src/service/storage"
	"github.com/zerops-io/zcli/src/service/sudoers"
)

type Config struct {
}

type RunConfig struct {
}

type Handler struct {
	config  Config
	logger  logger.Logger
	sudoers *sudoers.Handler
	storage *storage.Handler
}

func New(
	config Config,
	logger logger.Logger,
	sudoers *sudoers.Handler,
	storage *storage.Handler,
) *Handler {
	return &Handler{
		config:  config,
		logger:  logger,
		sudoers: sudoers,
		storage: storage,
	}
}

func (h *Handler) Run(_ context.Context, _ RunConfig) error {

	err := h.cleanVpn()
	if err != nil {
		return err
	}

	h.storage.Data.ProjectId = ""
	h.storage.Data.ServerIp = ""
	err = h.storage.Save()
	if err != nil {
		return err
	}

	h.logger.Info("\nvpn connection was closed\n")

	return nil
}
