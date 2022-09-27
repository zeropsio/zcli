package importProjectService

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
)

func (h *Handler) sendProjectRequest(ctx context.Context, config RunConfig, yamlContent string) ([]*zBusinessZeropsApiProtocol.ProjectImportServiceStack, error) {
	clientId, err := h.getClientId(ctx, config)
	if err != nil {
		return nil, err
	}

	res, err := h.apiGrpcClient.PostProjectImport(ctx, &zBusinessZeropsApiProtocol.PostProjectImportRequest{
		ClientId: clientId,
		Yaml:     yamlContent,
	})
	if err := proto.BusinessError(res, err); err != nil {
		return nil, err
	}

	fmt.Println(constants.Success + i18n.ProjectCreated + i18n.Success)

	return res.GetOutput().GetServiceStacks(), nil
}
