package startStopDelete

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) ServiceStart(ctx context.Context, projectId string, config RunConfig) error {

	serviceName := config.ServiceName
	fmt.Println(serviceName, projectId)

	fmt.Println(i18n.StartServiceProcessInit)
	// todo call api

	fmt.Println(i18n.StartServiceSuccess)

	return nil
}
