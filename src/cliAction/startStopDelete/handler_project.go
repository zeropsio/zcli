package startStopDelete

import (
	"context"

	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
)

func (h *Handler) ProjectStart(ctx context.Context, projectId string, _ string) (string, error) {
	startProjectResponse, err := h.apiGrpcClient.PutProjectStart(ctx, &zBusinessZeropsApiProtocol.PutProjectStartRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(startProjectResponse, err); err != nil {
		return "", err
	}

	return startProjectResponse.GetOutput().GetId(), nil
}

func (h *Handler) ProjectStop(ctx context.Context, projectId string, _ string) (string, error) {
	stopProjectResponse, err := h.apiGrpcClient.PutProjectStop(ctx, &zBusinessZeropsApiProtocol.PutProjectStopRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(stopProjectResponse, err); err != nil {
		return "", err
	}

	return stopProjectResponse.GetOutput().GetId(), nil
}

func (h *Handler) ProjectDelete(ctx context.Context, projectId string, _ string) (string, error) {
	deleteProjectResponse, err := h.apiGrpcClient.DeleteProject(ctx, &zBusinessZeropsApiProtocol.DeleteProjectRequest{
		Id: projectId,
	})

	if err := proto.BusinessError(deleteProjectResponse, err); err != nil {
		return "", err
	}

	return deleteProjectResponse.GetOutput().GetId(), nil
}
