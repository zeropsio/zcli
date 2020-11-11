package vpn

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
	"google.golang.org/grpc"
)

func (h *Handler) startVpnClient(ctx context.Context, targetAddress string) (_ zeropsVpnProtocol.ZeropsVpnProtocolClient, closeFunc func(), _ error) {

	h.logger.Debug("vpn client start")

	connection, err := grpc.DialContext(
		ctx,
		targetAddress+vpnApiGrpcPort,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			return invoker(timeoutCtx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	return zeropsVpnProtocol.NewZeropsVpnProtocolClient(connection), func() {
		_ = connection.Close()
	}, nil
}
