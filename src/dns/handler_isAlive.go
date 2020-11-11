package dns

import (
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/zerops-io/zcli/src/i18n"
)

func IsAlive() (bool, error) {

	_, err := exec.LookPath("dig")
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		ipAddress, err := exec.CommandContext(ctx, "dig", "+short", "core-master.zerops", "AAAA").Output()
		if err != nil {
			return false, err
		}
		if string(ipAddress) == "" {
			return false, nil
		}

		return true, nil
	}

	// ping6 fallback
	_, err = exec.LookPath("ping6")
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		_, err = exec.CommandContext(ctx, "ping6", "-c", "1", "core-master.zerops").Output()
		if err != nil {
			return false, nil
		}

		return true, nil
	}

	return false, errors.New(i18n.VpnStatusDnsNoCheckFunction)
}
