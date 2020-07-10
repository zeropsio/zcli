// +build linux

package startVpn

import (
	"errors"
	"os/exec"

	"github.com/zerops-io/zcli/src/service/sudoers"
)

func (h *Handler) cleanVpn() error {

	var err error

	_, err = h.sudoers.RunCommand(exec.Command("ip", "link", "del", "dev", "wg0"))
	if err != nil {
		if !errors.Is(err, sudoers.CannotFindDeviceErr) {
			return err
		}
	}

	return nil
}
