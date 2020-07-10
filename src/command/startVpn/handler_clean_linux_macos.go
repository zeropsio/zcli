// +build darwin

package startVpn

import "os/exec"

func (h *Handler) cleanVpn() error {

	var err error

	cmd := "ps aux | grep wireguard | grep -v grep | awk '{print $2}' | xargs sudo kill"

	_, err = h.sudoers.RunCommand(exec.Command("bash", "-c", cmd))
	if err != nil {
		return err
	}

	return nil
}
