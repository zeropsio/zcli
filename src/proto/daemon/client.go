package daemon

import (
	"context"

	"google.golang.org/grpc"
)

func CreateClient(ctx context.Context) (_ ZeropsDaemonProtocolClient, closeFunc func(), err error) {
	connection, err := grpc.DialContext(ctx, daemonDialAddress(), grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	closeFunc = func() { _ = connection.Close() }

	return NewZeropsDaemonProtocolClient(connection), closeFunc, nil
}
