package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetAllOrgs(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) ([]entity.Org, error) {
	response, err := restApiClient.GetUserInfo(ctx)
	if err != nil {
		return nil, err
	}

	resOutput, err := response.Output()
	if err != nil {
		return nil, err
	}

	orgs := make([]entity.Org, 0, len(resOutput.ClientUserList))
	for _, client := range resOutput.ClientUserList {
		orgs = append(orgs, orgFromEsSearch(client))
	}

	return orgs, nil
}

func GetOrgById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	clientId uuid.ClientId,
) (entity.Org, error) {
	orgs, err := GetAllOrgs(ctx, restApiClient)
	if err != nil {
		return entity.Org{}, err
	}
	org, found := gn.FindFirst(orgs, func(in entity.Org) bool {
		return in.ID == clientId
	})
	if !found {
		return entity.Org{}, errors.Errorf("Org [%s] not found", clientId)
	}
	return org, nil
}

func orgFromEsSearch(esClientUser output.ClientUserExtra) entity.Org {
	return entity.Org{
		ID:   esClientUser.ClientId,
		Name: esClientUser.Client.AccountName,
		Role: esClientUser.RoleCode,
	}
}
