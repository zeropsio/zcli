package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) ServiceStop(ctx context.Context, serviceId string) error {

	fmt.Println(i18n.StopServiceProcessInit)
	// todo call api
	// todo call api
	fmt.Println(serviceId)

	fmt.Println(i18n.StopServiceSuccess)

	return nil
}
