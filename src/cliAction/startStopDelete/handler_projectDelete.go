package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) ProjectDelete(ctx context.Context, projectId string, config RunConfig) error {

	if !config.Confirm {
		// run confirm dialogue
		shouldDelete := askForConfirmation(constants.Project)
		if !shouldDelete {
			fmt.Println(i18n.DelProjectCanceledByUser)
			return nil
		}
	}

	deleteProjectResponse, err := h.apiGrpcClient.DeleteProject(ctx, &business.DeleteProjectRequest{
		Id: projectId,
	})
	if err := proto.BusinessError(deleteProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.DeleteProjectProcessInit)

	processId := deleteProjectResponse.GetOutput().GetId()

	err = h.checkProcess(ctx, processId)
	if err != nil {
		return err
	}

	fmt.Println(i18n.DeleteProjectSuccess)

	return nil
}
