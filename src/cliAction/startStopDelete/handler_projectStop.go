package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) ProjectStop(ctx context.Context, projectId string) error {

	stopProjectResponse, err := h.apiGrpcClient.PutProjectStop(ctx, &business.PutProjectStopRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(stopProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.StopProjectProcessInit)

	processId := stopProjectResponse.GetOutput().GetId()

	err = h.checkProcess(ctx, processId)
	if err != nil {
		return err
	}

	fmt.Println(i18n.StopProjectSuccess)

	return nil
}
