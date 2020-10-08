package vpn

import (
	"context"

	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
	"google.golang.org/grpc"
)

func (h *Handler) startVpnClient(ctx context.Context, targetAddress string) (_ zeropsVpnProtocol.ZeropsVpnProtocolClient, closeFunc func(), _ error) {

	h.logger.Debug("vpn client start")

	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithInsecure())

	connection, err := grpc.DialContext(
		ctx,
		targetAddress+vpnApiGrpcPort,
		dialOpts...,
	)
	if err != nil {
		return nil, nil, err
	}

	return zeropsVpnProtocol.NewZeropsVpnProtocolClient(connection), func() {
		_ = connection.Close()
	}, nil
}
