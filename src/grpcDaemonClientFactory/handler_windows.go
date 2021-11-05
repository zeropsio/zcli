//go:build windows
// +build windows

package grpcDaemonClientFactory

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
	"google.golang.org/grpc"
)

func (h *Handler) CreateClient(ctx context.Context) (_ zeropsDaemonProtocol.ZeropsDaemonProtocolClient, closeFunc func(), err error) {
	connection, err := grpc.DialContext(ctx, "unix:///"+constants.SocketFilePath, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	closeFunc = func() { _ = connection.Close() }

	return zeropsDaemonProtocol.NewZeropsDaemonProtocolClient(connection), closeFunc, nil
}
