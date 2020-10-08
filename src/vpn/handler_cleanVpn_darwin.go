// +build darwin

package vpn

import (
	"os/exec"

	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

func (h *Handler) cleanVpn() error {

	var err error

	h.logger.Debug("clean vpn start")

	cmd := "ps aux | grep wireguard | grep -v grep | awk '{print $2}' | xargs sudo kill"
	_, err = cmdRunner.Run(exec.Command("bash", "-c", cmd))
	if err != nil {
		return err
	}

	h.logger.Debug("clean vpn end")

	return nil
}
