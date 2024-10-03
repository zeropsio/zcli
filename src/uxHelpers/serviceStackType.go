package uxHelpers

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

func PrintServiceStackTypeSelector(
	ctx context.Context,
	uxBlocks uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
) (result entity.ServiceStackType, _ error) {
	serviceStackTypes, err := repository.GetServiceStackTypes(ctx, restApiClient)
	if err != nil {
		return result, err
	}

	return SelectOne(ctx, uxBlocks, serviceStackTypes,
		SelectOneWithHeader[entity.ServiceStackType]("ID", "Name"),
		SelectOneWithSelectLabel[entity.ServiceStackType](i18n.T(i18n.ServiceStackTypeSelectorPrompt)),
		SelectOneWithNotFound[entity.ServiceStackType](i18n.T(i18n.ServiceStackTypeSelectorOutOfRangeError)),
		SelectOneWithRow(func(in entity.ServiceStackType) []string {
			return []string{string(in.ID), in.Name.String()}
		}),
	)
}
