package startStopDeleteProject

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) RunStop(ctx context.Context, config RunConfig, projectId string) error {

	stopProjectResponse, err := h.apiGrpcClient.PutProjectStop(ctx, &zeropsApiProtocol.PutProjectStopRequest{
		Id: projectId,
	})
	if err := utils.HandleGrpcApiError(stopProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.StopProjectProcessInit)

	processId := stopProjectResponse.GetOutput().GetId()

	// check process until FINISHED or CANCELED/FAILED
	err = h.checkProcess(ctx, processId)
	if err != nil {
		return err
	}

	fmt.Println(i18n.StopProcessSuccess)

	return nil
}
