package removeDaemon

import (
	"github.com/zerops-io/zcli/src/cliAction/stopVpn"
	"github.com/zerops-io/zcli/src/daemonInstaller"
)

type Config struct {
}

type RunConfig struct {
}

type Handler struct {
	config          Config
	daemonInstaller *daemonInstaller.Handler
	stopVpn         *stopVpn.Handler
}

func New(
	config Config,
	daemonInstaller *daemonInstaller.Handler,
	stopVpn *stopVpn.Handler,
) *Handler {
	return &Handler{
		config:          config,
		daemonInstaller: daemonInstaller,
		stopVpn:         stopVpn,
	}
}
