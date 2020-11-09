package vpn

import "context"

func (h *Handler) vpnStatusCheck(ctx context.Context) {
	h.lock.Lock()
	defer h.lock.Unlock()

	data := h.storage.Data()
	if data.VpnStarted {
		if !h.isVpnAlive(data.ServerIp) {
			err := h.startVpn(
				ctx,
				data.GrpcApiAddress,
				data.GrpcVpnAddress,
				data.Token,
				data.ProjectId,
				data.UserId,
				data.Mtu,
				data.CaCertificateUrl,
			)
			if err != nil {
				h.logger.Error(err)
			}
		}
	}
}
