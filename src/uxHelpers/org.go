package uxHelpers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

type orgSelectorConfig struct {
	skipOnOneItem bool
}
type OrgSelectorOption = gn.Option[orgSelectorConfig]

func WithOrgPickOnlyOneItem(b bool) OrgSelectorOption {
	return func(s *orgSelectorConfig) {
		s.skipOnOneItem = b
	}
}

func PrintOrgSelector(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	opts ...OrgSelectorOption,
) (entity.Org, error) {
	cfg := gn.ApplyOptions(opts...)

	orgs, err := repository.GetAllOrgs(ctx, restApiClient)
	if err != nil {
		return entity.Org{}, err
	}

	if len(orgs) == 0 {
		return entity.Org{}, errors.New(i18n.T(i18n.OrgSelectorListEmpty))
	}

	if len(orgs) == 1 && cfg.skipOnOneItem {
		return orgs[0], nil
	}

	header, body := createOrgTableRows(orgs)

	selected, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel(i18n.T(i18n.OrgSelectorPrompt)),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return entity.Org{}, err
	}

	if selected > len(orgs)-1 {
		return entity.Org{}, errors.New(i18n.T(i18n.OrgSelectorOutOfRangeError))
	}

	return orgs[selected], nil
}

func createOrgTableRows(projects []entity.Org) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("ID", "Name", "Role")

	body := table.NewBody()
	for _, project := range projects {
		body.AddStringsRow(
			string(project.ID),
			project.Name.String(),
			project.Role.Native(),
		)
	}

	return header, body
}
