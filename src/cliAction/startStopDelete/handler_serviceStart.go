package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) ServiceStart(ctx context.Context, serviceId string) error {

	fmt.Println(i18n.StartServiceProcessInit)

	// todo call api
	fmt.Println(serviceId)

	fmt.Println(i18n.StartServiceSuccess)

	return nil
}
