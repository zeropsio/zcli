package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) ProjectStop(ctx context.Context, projectId string) error {

	fmt.Println(i18n.StopProjectProcessInit)

	stopProjectResponse, err := h.apiGrpcClient.PutProjectStop(ctx, &business.PutProjectStopRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(stopProjectResponse, err); err != nil {
		return err
	}

	processId := stopProjectResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(constants.Success + i18n.StopProjectSuccess)

	return nil
}
