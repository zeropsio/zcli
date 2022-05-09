package vpn

import (
	"github.com/zerops-io/zcli/src/proto/daemon"
)

func (h *Handler) StopVpn() (_ *daemon.StopVpnResponse, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if !h.storage.Data().VpnStarted {
		return &daemon.StopVpnResponse{}, nil
	}
	vpnStatus, err := h.stopVpn()
	if err != nil {
		h.logger.Error(err)
		return nil, err
	}

	return &daemon.StopVpnResponse{
		VpnStatus:    vpnStatus,
		ActiveBefore: true,
	}, nil
}
