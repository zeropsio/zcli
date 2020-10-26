package installDaemon

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) Run(ctx context.Context, _ RunConfig) error {

	err := h.daemonInstaller.Install()
	if err != nil {
		return err
	}

	fmt.Println(i18n.DaemonInstallSuccess)

	return nil
}
