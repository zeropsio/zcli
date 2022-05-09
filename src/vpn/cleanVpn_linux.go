//go:build linux
// +build linux

package vpn

import (
	"errors"
	"os/exec"

	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

func (h *Handler) cleanVpn() error {

	var err error

	h.logger.Debug("clean vpn start")

	_, err = cmdRunner.Run(exec.Command("ip", "link", "del", "dev", "wg0"))
	if err != nil {
		if !errors.Is(err, cmdRunner.CannotFindDeviceErr) {
			return err
		}
	}

	h.logger.Debug("clean vpn end")

	return nil
}
