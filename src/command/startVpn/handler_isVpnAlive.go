package startVpn

import (
	"os/exec"
)

func (h *Handler) isVpnAlive() bool {

	if h.storage.Data.ServerIp == "" {
		return false
	}

	for i := 0; i < 3; i++ {
		_, err := exec.Command("ping6", h.storage.Data.ServerIp, "-c", "1").Output()
		if err != nil {
			continue
		}

		return true
	}

	return false
}
