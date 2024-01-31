package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/output"
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

func orgFromEsSearch(esClientUser output.ClientUserExtra) entity.Org {
	return entity.Org{
		ID:   esClientUser.ClientId,
		Name: esClientUser.Client.AccountName,
		Role: esClientUser.RoleCode,
	}
}
