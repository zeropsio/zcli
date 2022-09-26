package vpn

import (
	"context"
)

func (h *Handler) StartVpn(
	ctx context.Context,
	grpcApiAddress string,
	grpcVpnAddress string,
	token string,
	projectId string,
	userId string,
	mtu uint32,
	caCertificateUrl string,
	preferredPortMin uint32,
	preferredPortMax uint32,
) (err error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.startVpn(
		ctx,
		grpcApiAddress,
		grpcVpnAddress,
		token,
		projectId,
		userId,
		mtu,
		caCertificateUrl,
		preferredPortMin,
		preferredPortMax,
	)
}
