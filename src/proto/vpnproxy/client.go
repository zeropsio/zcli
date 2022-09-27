package vpnproxy

import (
	"context"

	"github.com/zeropsio/zcli/src/proto/unary"

	"google.golang.org/grpc"
)

func CreateClient(ctx context.Context, targetAddress string) (_ ZeropsVpnProtocolClient, closeFunc func(), _ error) {
	conn, err := grpc.DialContext(
		ctx,
		targetAddress,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(unary.TimeoutInterceptor, unary.SupportInterceptor(IsInternal)),
	)
	if err != nil {
		return nil, nil, err
	}

	return NewZeropsVpnProtocolClient(conn), func() {
		_ = conn.Close()
	}, nil
}
