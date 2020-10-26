package installDaemon

import (
	"github.com/zerops-io/zcli/src/daemonInstaller"
)

type Config struct {
}

type RunConfig struct {
}

type Handler struct {
	config          Config
	daemonInstaller *daemonInstaller.Handler
}

func New(
	config Config,
	daemonInstaller *daemonInstaller.Handler,
) *Handler {
	return &Handler{
		config:          config,
		daemonInstaller: daemonInstaller,
	}
}
