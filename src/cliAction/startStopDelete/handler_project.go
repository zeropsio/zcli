package startStopDelete

import (
	"context"

	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func ProjectStart(ctx context.Context, h *Handler, projectId string, _ string) (string, error) {
	startProjectResponse, err := h.apiGrpcClient.PutProjectStart(ctx, &business.PutProjectStartRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(startProjectResponse, err); err != nil {
		return "", err
	}

	return startProjectResponse.GetOutput().GetId(), nil
}

func ProjectStop(ctx context.Context, h *Handler, projectId string, _ string) (string, error) {
	stopProjectResponse, err := h.apiGrpcClient.PutProjectStop(ctx, &business.PutProjectStopRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(stopProjectResponse, err); err != nil {
		return "", err
	}

	return stopProjectResponse.GetOutput().GetId(), nil
}

func ProjectDelete(ctx context.Context, h *Handler, projectId string, _ string) (string, error) {
	deleteProjectResponse, err := h.apiGrpcClient.DeleteProject(ctx, &business.DeleteProjectRequest{
		Id: projectId,
	})

	if err := proto.BusinessError(deleteProjectResponse, err); err != nil {
		return "", err
	}

	return deleteProjectResponse.GetOutput().GetId(), nil
}
