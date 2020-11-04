package vpn

import (
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
)

func (h *Handler) StopVpn() (vpnStatus *zeropsDaemonProtocol.VpnStatus, err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	vpnStatus, err = h.stopVpn()
	if err != nil {
		h.logger.Error(err)
		return nil, err
	}

	return vpnStatus, nil
}
