package vpn

func (h *Handler) StatusVpn() bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	data := h.storage.Data()
	return h.isVpnAlive(data.ServerIp)
}
