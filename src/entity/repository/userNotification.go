package repository

import (
	"context"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// GetUserNotificationsByProject fetches notifications for a project with pagination,
// sorted by actionCreated descending.
func GetUserNotificationsByProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
	limit int,
	offset int,
) ([]entity.UserNotification, error) {
	esFilter := body.EsFilter{
		Search: body.EsFilterSearch{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    orgId.TypedString(),
			},
			{
				Name:     "project.id",
				Operator: "eq",
				Value:    projectId.TypedString(),
			},
		},
		Sort: body.EsFilterSort{
			{
				Name:      "actionCreated",
				Ascending: types.NewBoolNull(false),
			},
		},
		Limit:  types.NewIntNull(limit),
		Offset: types.NewIntNull(offset),
	}

	search, err := restApiClient.PostUserNotificationSearch(ctx, esFilter)
	if err != nil {
		return nil, err
	}

	response, err := search.Output()
	if err != nil {
		return nil, err
	}

	return gn.TransformSlice(response.Items, userNotificationFromEsSearch), nil
}

func userNotificationFromEsSearch(esNotification output.EsUserNotification) entity.UserNotification {
	serviceNames := make([]string, 0, len(esNotification.ServiceStacks))
	for _, ss := range esNotification.ServiceStacks {
		serviceNames = append(serviceNames, ss.Name.String())
	}

	var createdByUser string
	if email, ok := esNotification.CreatedByUser.Email.Get(); ok {
		createdByUser = email.Native()
	}
	if fullName, ok := esNotification.CreatedByUser.FullName.Get(); ok && fullName.Native() != "" {
		createdByUser = fullName.Native()
	}

	var projectId uuid.ProjectIdNull
	var projectName types.StringNull
	if esNotification.Project != nil {
		projectId = uuid.NewProjectIdNull(esNotification.Project.Id)
		projectName = types.NewStringNull(esNotification.Project.Name.String())
	}

	var errorMessage types.StringNull
	if esNotification.Error != nil {
		errorMessage = types.NewStringNull(esNotification.Error.Message.String())
	}

	return entity.UserNotification{
		Id:             esNotification.Id,
		OrgId:          esNotification.ClientId,
		ProjectId:      projectId,
		ProjectName:    projectName,
		Type:           esNotification.Type,
		ActionName:     esNotification.ActionName,
		ActionCreated:  esNotification.ActionCreated,
		ActionFinished: esNotification.ActionFinished,
		Acknowledged:   esNotification.Acknowledged,
		CreatedByUser:  createdByUser,
		ServiceNames:   serviceNames,
		ErrorMessage:   errorMessage,
	}
}
