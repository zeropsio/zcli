package repository

import (
	"testing"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// These pin the API-output -> entity field mapping before the domain-type swap.
// After the swap the boundary lives here, so the same fixtures must still yield
// the same underlying values (assertions move from zerops-go types to the new
// zcli domain types, but the data carried across must not change).

func TestProjectFromApiOutput(t *testing.T) {
	got := projectFromApiOutput(output.Project{
		Id:          uuid.ProjectId("proj-1"),
		ClientId:    uuid.ClientId("org-1"),
		Name:        types.String("proj-name"),
		Mode:        enum.ProjectModeEnum("HA"),
		Description: types.NewTextNull("a description"),
		Status:      enum.ProjectStatusEnum("ACTIVE"),
	})

	want := entity.Project{
		Id:          uuid.ProjectId("proj-1"),
		Name:        types.String("proj-name"),
		Mode:        enum.ProjectModeEnum("HA"),
		OrgId:       uuid.ClientId("org-1"),
		Description: types.Text("a description"),
		Status:      enum.ProjectStatusEnum("ACTIVE"),
	}
	if got != want {
		t.Errorf("projectFromApiOutput = %+v, want %+v", got, want)
	}
}

func TestProjectFromApiOutputEmptyDescription(t *testing.T) {
	got := projectFromApiOutput(output.Project{
		Id:   uuid.ProjectId("proj-1"),
		Name: types.String("proj-name"),
	})
	if got.Description != "" {
		t.Errorf("empty description should map to zero value, got %q", got.Description)
	}
}

func TestOrgFromEsSearch(t *testing.T) {
	got := orgFromEsSearch(output.ClientUserExtraWithClientLight{
		ClientId: uuid.ClientId("org-1"),
		Status:   enum.ClientUserStatusEnum("ACTIVE"),
		RoleCode: enum.ClientUserRoleCodeEnum("ADMIN"),
		Client: output.ClientLight{
			AccountName: types.String("acme"),
		},
	})

	want := entity.Org{
		Id:     uuid.ClientId("org-1"),
		Name:   types.String("acme"),
		Status: enum.ClientUserStatusEnum("ACTIVE"),
		Role:   enum.ClientUserRoleCodeEnum("ADMIN"),
	}
	if got != want {
		t.Errorf("orgFromEsSearch = %+v, want %+v", got, want)
	}
}

func TestProcessFromEsSearch(t *testing.T) {
	got := processFromEsSearch(output.EsProcess{
		Id:             uuid.ProcessId("proc-1"),
		ClientId:       uuid.ClientId("org-1"),
		ProjectId:      uuid.ProjectId("proj-1"),
		ServiceStackId: uuid.NewServiceStackIdNull(uuid.ServiceStackId("svc-1")),
		ActionName:     types.String("stack.start"),
		Status:         enum.ProcessStatusEnumRunning,
		Sequence:       types.Int(3),
	})

	if got.Id != uuid.ProcessId("proc-1") ||
		got.OrgId != uuid.ClientId("org-1") ||
		got.ProjectId != uuid.ProjectId("proj-1") ||
		got.ActionName != types.String("stack.start") ||
		got.Status != enum.ProcessStatusEnumRunning ||
		got.Sequence != types.Int(3) {
		t.Errorf("processFromEsSearch mismatch: %+v", got)
	}
	if svc, ok := got.ServiceId.Get(); !ok || svc != uuid.ServiceStackId("svc-1") {
		t.Errorf("processFromEsSearch ServiceId = %v (ok=%v), want svc-1", svc, ok)
	}
}
