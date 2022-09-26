package vpn

import "context"

func (h *Handler) StopVpn(ctx context.Context) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.stopVpn(ctx)
}
