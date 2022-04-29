package startStopDeleteProject

import (
	"context"
	"fmt"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) RunDelete(ctx context.Context, config RunConfig, projectId string, actionType string) error {

	if !config.Confirm {
		// run confirm dialogue
		shouldDelete := askForConfirmation()
		if !shouldDelete {
			fmt.Println(i18n.CanceledByUser)
			return nil
		}
	}

	deleteProjectResponse, err := h.apiGrpcClient.DeleteProject(ctx, &zeropsApiProtocol.DeleteProjectRequest{
		Id: projectId,
	})
	if err := utils.HandleGrpcApiError(deleteProjectResponse, err); err != nil {
		return err
	}

	fmt.Println(i18n.DeleteProjectProcessInit)

	processId := deleteProjectResponse.GetOutput().GetId()

	err = h.checkProcess(ctx, processId)
	if err != nil {
		return err
	}

	fmt.Println(i18n.DeleteProcessSuccess)

	return nil
}

func askForConfirmation() bool {
	fmt.Print(i18n.ConfirmDelete)
	var response string

	_, err := fmt.Scan(&response)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp := strings.ToLower(response)
	if resp == "y" || resp == "yes" {
		return true
	} else if resp == "n" || resp == "no" {
		return false
	} else {
		return askForConfirmation()
	}
}
