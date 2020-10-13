package vpn

import "context"

func (h *Handler) checkStatus(ctx context.Context) {
	h.lock.Lock()
	defer h.lock.Unlock()

	data := h.storage.Data()

	if data.ProjectId != "" {
		if !h.isVpnAlive(data.ServerIp) {
			err := h.startVpn(
				ctx,
				data.GrpcApiAddress,
				data.GrpcVpnAddress,
				data.Token,
				data.ProjectId,
				data.Mtu,
			)
			if err != nil {
				h.logger.Error(err)
			}
		}
	}
}
