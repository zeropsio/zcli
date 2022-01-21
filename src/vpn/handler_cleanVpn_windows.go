//go:build windows
// +build windows

package vpn

import "os/exec"

func (h *Handler) cleanVpn() error {
	output, err := exec.Command("wireguard", "/uninstalltunnelservice", "zerops").Output()
	if err != nil {
		h.logger.Error(output)
		return err
	}
	return nil
}
