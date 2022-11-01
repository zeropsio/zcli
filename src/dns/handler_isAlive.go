package dns

import (
	"context"
	"errors"
	"time"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/nettools"
)

func IsAlive() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if !nettools.HasIPv6PingCommand() {
		return false, errors.New(i18n.VpnStatusDnsNoCheckFunction)
	}
	err := nettools.Ping(ctx, "node1.master.core.zerops")
	if err != nil {
		return false, nil
	}
	return true, nil
}
