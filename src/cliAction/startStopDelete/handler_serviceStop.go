package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) ServiceStop(ctx context.Context, serviceId string) error {
	stopServiceResponse, err := h.apiGrpcClient.PutServiceStackStop(ctx, &business.PutServiceStackStopRequest{
		Id: serviceId,
	})
	if err := proto.BusinessError(stopServiceResponse, err); err != nil {
		return err
	}
	fmt.Println(i18n.StopServiceProcessInit)
	processId := stopServiceResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(i18n.StopServiceSuccess)

	return nil
}
