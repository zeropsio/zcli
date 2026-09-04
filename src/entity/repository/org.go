package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func GetAllOrgs(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
) ([]entity.Org, error) {
	var orgs []entity.Org
	for offset := 0; ; {
		response, err := restApiClient.GetUserClientList(ctx, query.ListUserClients{
			Limit:  types.NewIntNull(listPageLimit),
			Offset: types.NewIntNull(offset),
		})
		if err != nil {
			return nil, err
		}

		resOutput, err := response.Output()
		if err != nil {
			return nil, err
		}

		for _, client := range resOutput.List {
			orgs = append(orgs, orgFromClientUser(client))
		}

		offset += len(resOutput.List)
		if len(resOutput.List) == 0 || offset >= resOutput.Total.Native() {
			break
		}
	}

	return orgs, nil
}

func GetOrgById(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId uuid.ClientId,
) (entity.Org, error) {
	orgs, err := GetAllOrgs(ctx, restApiClient)
	if err != nil {
		return entity.Org{}, err
	}
	org, found := gn.FindFirst(orgs, func(in entity.Org) bool {
		return in.Id == orgId
	})
	if !found {
		return entity.Org{}, errors.Errorf("Org [%s] not found", orgId)
	}
	return org, nil
}

func orgFromClientUser(clientUser output.ClientUserExtraWithClientLight) entity.Org {
	return entity.Org{
		Id:     clientUser.ClientId,
		Name:   clientUser.Client.AccountName,
		Status: clientUser.Status,
		Role:   clientUser.RoleCode,
	}
}
