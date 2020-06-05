package stopVpn

import (
	"context"
	"errors"
	"os/exec"

	"github.com/zerops-io/zcli/src/service/storage"

	"github.com/zerops-io/zcli/src/service/logger"
	"github.com/zerops-io/zcli/src/service/sudoers"
)

type Config struct {
}

type RunConfig struct {
}

type Handler struct {
	config  Config
	logger  *logger.Handler
	sudoers *sudoers.Handler
	storage *storage.Handler
}

func New(
	config Config,
	logger *logger.Handler,
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

	_, err := h.sudoers.RunCommand(exec.Command("ip", "link", "del", "dev", "wg0"))
	if err != nil {
		if !errors.Is(err, sudoers.CannotFindDeviceErr) {
			return err
		}
	}

	h.storage.Data.ProjectId = ""
	h.storage.Data.ServerIp = ""
	err = h.storage.Save()
	if err != nil {
		return err
	}

	return nil
}
