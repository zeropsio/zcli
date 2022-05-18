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

func (h *Handler) ServiceDelete(ctx context.Context, serviceId string, config RunConfig) error {

	if !config.Confirm {
		// run confirm dialogue
		shouldDelete := askForConfirmation(constants.Service)
		if !shouldDelete {
			fmt.Println(i18n.DelServiceCanceledByUser)
			return nil
		}
	}

	fmt.Println(i18n.DeleteServiceProcessInit)

	deleteServiceResponse, err := h.apiGrpcClient.DeleteServiceStack(ctx, &business.DeleteServiceStackRequest{
		Id: serviceId,
	})
	if err := proto.BusinessError(deleteServiceResponse, err); err != nil {
		return err
	}

	processId := deleteServiceResponse.GetOutput().GetId()

	err = processChecker.CheckProcess(ctx, processId, h.apiGrpcClient)
	if err != nil {
		return err
	}

	fmt.Println(constants.Success + i18n.DeleteServiceSuccess)

	return nil
}
