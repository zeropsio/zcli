package startStopDelete

import (
	"context"

	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
)

func (h *Handler) ServiceStart(ctx context.Context, _ string, serviceId string) (string, error) {
	startServiceResponse, err := h.apiGrpcClient.PutServiceStackStart(ctx, &zBusinessZeropsApiProtocol.PutServiceStackStartRequest{
		Id: serviceId,
	})
	if err := proto.BusinessError(startServiceResponse, err); err != nil {
		return "", err
	}

	return startServiceResponse.GetOutput().GetId(), nil
}

func (h *Handler) ServiceStop(ctx context.Context, _ string, serviceId string) (string, error) {
	stopServiceResponse, err := h.apiGrpcClient.PutServiceStackStop(ctx, &zBusinessZeropsApiProtocol.PutServiceStackStopRequest{
		Id: serviceId,
	})
	if err := proto.BusinessError(stopServiceResponse, err); err != nil {
		return "", err
	}

	return stopServiceResponse.GetOutput().GetId(), nil
}

func (h *Handler) ServiceDelete(ctx context.Context, _ string, serviceId string) (string, error) {
	deleteServiceResponse, err := h.apiGrpcClient.DeleteServiceStack(ctx, &zBusinessZeropsApiProtocol.DeleteServiceStackRequest{
		Id: serviceId,
	})
	if err := proto.BusinessError(deleteServiceResponse, err); err != nil {
		return "", err
	}

	return deleteServiceResponse.GetOutput().GetId(), nil
}
