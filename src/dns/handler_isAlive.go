package dns

import (
	"context"
	"errors"
	"time"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/nettools"
)

func IsAlive() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if !nettools.HasIPv6PingCommand() {
		return false, errors.New(i18n.VpnStatusDnsNoCheckFunction)
	}
	err := nettools.Ping(ctx, "core-master")
	if err != nil {
		return false, nil
	}
	return true, nil
}
