package removeDaemon

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeropsio/zcli/src/daemonInstaller"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeropsio/zcli/src/cliAction/stopVpn"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) Run(ctx context.Context, _ RunConfig) error {

	if h.daemonInstaller.IsInstalled() {
		err := h.stopVpn.Run(ctx, stopVpn.RunConfig{})
		if err != nil {
			if errStatus, ok := status.FromError(err); ok {
				if errStatus.Code() == codes.Unavailable {
					fmt.Println(i18n.DaemonRemoveStopVpnUnavailable)
				}
			}
		}
	}

	err := h.daemonInstaller.Remove()
	if errors.Is(err, daemonInstaller.ErrElevatedPrivileges) {
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println(i18n.DaemonRemoveSuccess)

	return nil
}
