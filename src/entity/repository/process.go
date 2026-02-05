package repository

import (
	"context"
	"slices"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

const maxProcessSearchResults = 100

func GetProcessByActionNameAndProjectId(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
	actionName types.String,
) ([]entity.Process, error) {
	search, err := restApiClient.PostProcessSearch(ctx, body.EsFilter{
		Search: body.EsFilterSearch{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    orgId.TypedString(),
			},
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    projectId.TypedString(),
			},
			{
				Name:     "actionName",
				Operator: "eq",
				Value:    actionName,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	response, err := search.Output()
	if err != nil {
		return nil, err
	}
	return gn.TransformSlice(response.Items, processFromEsSearch), nil
}

func processFromEsSearch(esProcess output.EsProcess) entity.Process {
	return entity.Process{
		Id:         esProcess.Id,
		OrgId:      esProcess.ClientId,
		ProjectId:  esProcess.ProjectId,
		ServiceId:  esProcess.ServiceStackId,
		ActionName: esProcess.ActionName,
		Status:     esProcess.Status,
		Sequence:   esProcess.Sequence,
	}
}

func processFromApiOutput(process output.Process) entity.Process {
	return entity.Process{
		Id:         process.Id,
		OrgId:      process.ClientId,
		ProjectId:  process.ProjectId,
		ServiceId:  process.ServiceStackId,
		ActionName: process.ActionName,
		Status:     process.Status,
		Sequence:   process.Sequence,
	}
}

// GetRunningAndPendingProcessesByProject fetches RUNNING and PENDING processes
// for a specific project, sorted by creation date descending.
func GetRunningAndPendingProcessesByProject(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
) ([]entity.Process, error) {
	// EsFilter doesn't support OR conditions, so we make two calls and merge.
	// Both calls return sorted results, but we re-sort after merge for correctness.
	pendingProcesses, err := getProcessesByStatus(ctx, restApiClient, orgId, projectId, enum.ProcessStatusEnumPending)
	if err != nil {
		return nil, err
	}

	runningProcesses, err := getProcessesByStatus(ctx, restApiClient, orgId, projectId, enum.ProcessStatusEnumRunning)
	if err != nil {
		return nil, err
	}

	allProcesses := make([]entity.Process, 0, len(pendingProcesses)+len(runningProcesses))
	allProcesses = append(allProcesses, pendingProcesses...)
	allProcesses = append(allProcesses, runningProcesses...)

	slices.SortFunc(allProcesses, func(a, b entity.Process) int {
		if a.Created.Native().After(b.Created.Native()) {
			return -1
		}
		if a.Created.Native().Before(b.Created.Native()) {
			return 1
		}
		return 0
	})

	if len(allProcesses) > maxProcessSearchResults {
		allProcesses = allProcesses[:maxProcessSearchResults]
	}

	return allProcesses, nil
}

func getProcessesByStatus(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	orgId uuid.ClientId,
	projectId uuid.ProjectId,
	status enum.ProcessStatusEnum,
) ([]entity.Process, error) {
	esFilter := body.EsFilter{
		Search: body.EsFilterSearch{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    orgId.TypedString(),
			},
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    projectId.TypedString(),
			},
			{
				Name:     "status",
				Operator: "eq",
				Value:    types.NewString(status.String()),
			},
		},
		Sort: body.EsFilterSort{
			{
				Name:      "created",
				Ascending: types.NewBoolNull(false),
			},
		},
		Limit: types.NewIntNull(maxProcessSearchResults),
	}

	search, err := restApiClient.PostProcessSearch(ctx, esFilter)
	if err != nil {
		return nil, err
	}

	response, err := search.Output()
	if err != nil {
		return nil, err
	}

	return gn.TransformSlice(response.Items, processFromEsSearchExtended), nil
}

func processFromEsSearchExtended(esProcess output.EsProcess) entity.Process {
	serviceNames := make([]string, 0, len(esProcess.ServiceStacks))
	for _, ss := range esProcess.ServiceStacks {
		serviceNames = append(serviceNames, ss.Name.String())
	}

	var createdByUser string
	if email, ok := esProcess.CreatedByUser.Email.Get(); ok {
		createdByUser = email.Native()
	}
	if fullName, ok := esProcess.CreatedByUser.FullName.Get(); ok && fullName.Native() != "" {
		createdByUser = fullName.Native()
	}

	return entity.Process{
		Id:              esProcess.Id,
		OrgId:           esProcess.ClientId,
		ProjectId:       esProcess.ProjectId,
		ServiceId:       esProcess.ServiceStackId,
		ActionName:      esProcess.ActionName,
		Status:          esProcess.Status,
		Sequence:        esProcess.Sequence,
		Created:         esProcess.Created,
		LastUpdate:      esProcess.LastUpdate,
		Started:         esProcess.Started,
		CreatedByUser:   createdByUser,
		ServiceNames:    serviceNames,
		CreatedBySystem: esProcess.CreatedBySystem,
	}
}
