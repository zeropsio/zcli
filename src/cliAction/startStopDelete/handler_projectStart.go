package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) ProjectStart(ctx context.Context, projectId string) error {

	fmt.Println(i18n.StartProjectProcessInit)

	startProjectResponse, err := h.apiGrpcClient.PutProjectStart(ctx, &business.PutProjectStartRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(startProjectResponse, err); err != nil {
		return err
	}

	processId := startProjectResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println("✓ " + i18n.StartProjectSuccess)

	return nil
}
