package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
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
	// todo call api
	fmt.Println(serviceId)

	fmt.Println(i18n.DeleteServiceSuccess)

	return nil
}
