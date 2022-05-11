package startStopDelete

import (
	"context"
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) getServiceId(ctx context.Context, config RunConfig, projectId string) (string, error) {

	if config.ServiceName == "" {
		return "", errors.New(i18n.ServiceNameIsEmpty)
	}

	response, err := h.apiGrpcClient.GetServiceStackByName(ctx, &business.GetServiceStackByNameRequest{
		ProjectId: projectId,
		Name:      config.ServiceName,
	})
	if err := proto.BusinessError(response, err); err != nil {
		return "", err
	}

	id := response.GetOutput().GetId()
	fmt.Println(id)

	// TODO check response
	if len(id) == 0 {
		return "", errors.New(i18n.ServiceNotFound)
	}

	return id, nil
}
