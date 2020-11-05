package vpn

import (
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StopVpn() (_ *zeropsDaemonProtocol.StopVpnResponse, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if !h.storage.Data().VpnStarted {
		return &zeropsDaemonProtocol.StopVpnResponse{}, nil
	}
	vpnStatus, err := h.stopVpn()
	if err != nil {
		h.logger.Error(err)
		return nil, err
	}

	return &zeropsDaemonProtocol.StopVpnResponse{
		VpnStatus:    vpnStatus,
		ActiveBefore: true,
	}, nil
}
