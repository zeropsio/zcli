package installDaemon

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeropsio/zcli/src/daemonInstaller"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) Run(ctx context.Context, _ RunConfig) error {

	err := h.daemonInstaller.Install()
	if errors.Is(err, daemonInstaller.ErrElevatedPrivileges) {
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println(i18n.DaemonInstallSuccess)

	return nil
}
