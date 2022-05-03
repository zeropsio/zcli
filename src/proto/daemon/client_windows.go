//go:build windows
// +build windows

package daemon

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
	"google.golang.org/grpc"
)

func CreateClient(ctx context.Context) (_ ZeropsDaemonProtocolClient, closeFunc func(), err error) {
	connection, err := grpc.DialContext(ctx, "localhost"+constants.DaemonAddress, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	closeFunc = func() { _ = connection.Close() }

	return NewZeropsDaemonProtocolClient(connection), closeFunc, nil
}
