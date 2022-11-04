package vpn

import (
	"context"
	"fmt"
	"os"
)

func (h *Handler) dnsCleanNetworkSetup(ctx context.Context) error {

	os.Remove("/etc/resolver/zerops")

	stdin := fmt.Sprintf(`remove State:/Network/Service/zerops_vpn_service/IPv6
remove Setup:/Network/Service/zerops_vpn_service/IPv6`)

	if _, err := h.runCommand(ctx, makeCommand("scutil", commandWithStdin(stdin))); err != nil {
		return err
	}

	return nil
}
