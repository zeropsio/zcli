package vpn

import (
	"context"
	"os"
)

func (h *Handler) dnsCleanNetworkSetup(ctx context.Context) error {

	os.Remove("/etc/resolver/zerops")
	{
		stdin := "remove State:/Network/Service/zerops_vpn_service/IPv6\n"
		if _, err := h.runCommand(ctx, makeCommand("scutil", commandWithStdin(stdin))); err != nil {
			return err
		}
	}

	{
		stdin := "remove Setup:/Network/Service/zerops_vpn_service/IPv6\n"
		if _, err := h.runCommand(ctx, makeCommand("scutil", commandWithStdin(stdin))); err != nil {
			return err
		}
	}

	return nil
}
