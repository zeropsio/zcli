package unary

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeropsio/zcli/src/utils/uuid"
	"google.golang.org/grpc/metadata"

	"github.com/zeropsio/zcli/src/support"

	"google.golang.org/grpc"
)

func TimeoutInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	return invoker(timeoutCtx, method, req, reply, cc, opts...)
}

func SupportInterceptor(isInternal func(any) bool) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		supportID, ok := support.GetID(ctx)
		if !ok {
			supportID = uuid.GetShort()
		}
		ctx = metadata.AppendToOutgoingContext(ctx, support.Key, supportID)
		err := invoker(ctx, method, req, res, cc, opts...)
		code := status.Code(err)
		if code == codes.Unknown || code == codes.Unavailable || isInternal(res) {
			fmt.Println("support id: ", supportID)
			return err
		}
		return err
	}
}
