package startStopDeleteProject

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/processChecker"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) RunStart(ctx context.Context, config RunConfig, projectId string) error {

	startProjectResponse, err := h.apiGrpcClient.PutProjectStart(ctx, &business.PutProjectStartRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(startProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.StartProjectProcessInit)

	processId := startProjectResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(i18n.StartProcessSuccess)

	return nil
}
