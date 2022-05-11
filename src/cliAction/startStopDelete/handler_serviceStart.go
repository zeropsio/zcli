package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) ServiceStart(ctx context.Context, serviceId string) error {

	startServiceResponse, err := h.apiGrpcClient.PutServiceStackStart(ctx, &business.PutServiceStackStartRequest{
		Id: serviceId,
	})
	if err := proto.BusinessError(startServiceResponse, err); err != nil {
		return err
	}
	fmt.Println(i18n.StartServiceProcessInit)
	processId := startServiceResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(i18n.StartServiceSuccess)

	return nil
}
