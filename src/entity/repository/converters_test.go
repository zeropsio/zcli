package repository

import (
	"testing"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// These pin the API-output -> domain-entity field mapping at the conversion
// boundary: the same zerops-go fixtures must yield the expected zcli domain values.

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
		Id:          entity.ProjectId("proj-1"),
		Name:        "proj-name",
		Mode:        entity.ProjectMode("HA"),
		OrgId:       entity.OrgId("org-1"),
		Description: "a description",
		Status:      entity.ProjectStatus("ACTIVE"),
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

func TestOrgFromClientUser(t *testing.T) {
	got := orgFromClientUser(output.ClientUserExtraWithClientLight{
		ClientId: uuid.ClientId("org-1"),
		Status:   enum.ClientUserStatusEnum("ACTIVE"),
		RoleCode: enum.ClientUserRoleCodeEnum("ADMIN"),
		Client: output.ClientLight{
			AccountName: types.String("acme"),
		},
	})

	want := entity.Org{
		Id:     entity.OrgId("org-1"),
		Name:   "acme",
		Status: entity.OrgStatus("ACTIVE"),
		Role:   entity.OrgRole("ADMIN"),
	}
	if got != want {
		t.Errorf("orgFromClientUser = %+v, want %+v", got, want)
	}
	if !got.Status.IsActive() {
		t.Error("ACTIVE org status should report IsActive")
	}
}

func TestProcessFromApiOutput(t *testing.T) {
	got := ProcessFromApiOutput(output.Process{
		Id:             uuid.ProcessId("proc-1"),
		ClientId:       uuid.ClientId("org-1"),
		ProjectId:      uuid.ProjectId("proj-1"),
		ServiceStackId: uuid.NewServiceStackIdNull(uuid.ServiceStackId("svc-1")),
		ActionName:     types.String("stack.start"),
		Status:         enum.ProcessStatusEnumRunning,
		Sequence:       types.Int(3),
	})

	if got.Id != entity.ProcessId("proc-1") ||
		got.OrgId != entity.OrgId("org-1") ||
		got.ProjectId != entity.ProjectId("proj-1") ||
		got.ActionName != "stack.start" ||
		got.Status != entity.ProcessStatus("RUNNING") ||
		got.Sequence != 3 {
		t.Errorf("processFromApiOutput mismatch: %+v", got)
	}
	if svc, ok := got.ServiceId.Get(); !ok || svc != entity.ServiceId("svc-1") {
		t.Errorf("processFromApiOutput ServiceId = %v (ok=%v), want svc-1", svc, ok)
	}
}
