package vpn

import (
	"context"
	"fmt"
	"os"
)

func (h *Handler) setDnsNetworksetup(ctx context.Context) error {

	if err := os.MkdirAll("/etc/resolver", 0755); err != nil {
		return err
	}

	data := h.storage.Data()
	if err := os.WriteFile(
		"/etc/resolver/zerops",
		[]byte(
			fmt.Sprintf("nameserver %s\nsearch zerops\n", data.DnsIp.String()),
		),
		0644,
	); err != nil {
		return err
	}

	{
		stdin := fmt.Sprintf(`d.init
d.add Addresses * fe80::1d04:6b6d:7ad7:85e4 2600:3c03::de:d002
d.add DestAddresses * ::ffff:ffff:ffff:ffff:0:0 ::
d.add Flags * 0 0
d.add InterfaceName %s
d.add PrefixLength * 64 116
set State:/Network/Service/zerops_vpn_service/IPv6
`, data.InterfaceName)

		if _, err := h.runCommand(ctx, makeCommand("scutil", commandWithStdin(stdin))); err != nil {
			return err
		}

	}
	{
		stdin := fmt.Sprintf(`d.init
d.add Addresses * fe80::1d04:6b6d:7ad7:85e4 2600:3c03::de:d002
d.add DestAddresses * ::ffff:ffff:ffff:ffff:0:0 ::
d.add Flags * 0 0
d.add InterfaceName %s
d.add PrefixLength * 64 116
set Setup:/Network/Service/zerops_vpn_service/IPv6
`, data.InterfaceName)

		if _, err := h.runCommand(ctx, makeCommand("scutil", commandWithStdin(stdin))); err != nil {
			return err
		}

	}
	return nil
}
