package vpnproxy

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

func CreateClient(ctx context.Context, targetAddress string) (_ ZeropsVpnProtocolClient, closeFunc func(), _ error) {
	conn, err := grpc.DialContext(
		ctx,
		targetAddress,
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

	return NewZeropsVpnProtocolClient(conn), func() {
		_ = conn.Close()
	}, nil
}
