package startStopDeleteProject

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) RunStart(ctx context.Context, config RunConfig, projectId string) error {

	startProjectResponse, err := h.apiGrpcClient.PutProjectStart(ctx, &zeropsApiProtocol.PutProjectStartRequest{
		Id: projectId,
	})
	if err := utils.HandleGrpcApiError(startProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.StartProjectProcessInit)

	processId := startProjectResponse.GetOutput().GetId()

	// check process until FINISHED or CANCELED/FAILED
	err = h.checkProcess(ctx, processId)
	if err != nil {
		return err
	}

	fmt.Println(i18n.StartProcessSuccess)

	return nil
}
