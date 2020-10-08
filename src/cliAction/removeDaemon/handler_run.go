package removeDaemon

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zerops-io/zcli/src/cliAction/stopVpn"

	"github.com/zerops-io/zcli/src/i18n"
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
	if err != nil {
		return err
	}

	fmt.Println(i18n.DaemonRemoveSuccess)

	return nil
}
