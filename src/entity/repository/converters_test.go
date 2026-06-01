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
		t.Errorf("ProcessFromApiOutput mismatch: %+v", got)
	}
	if svc, ok := got.ServiceId.Get(); !ok || svc != entity.ServiceId("svc-1") {
		t.Errorf("ProcessFromApiOutput ServiceId = %v (ok=%v), want svc-1", svc, ok)
	}
}

func TestAppVersionFromApiOutput(t *testing.T) {
	got := appVersionFromApiOutput(output.GetAppVersion{
		Id:             uuid.AppVersionId("app-1"),
		ClientId:       uuid.ClientId("org-1"),
		ProjectId:      uuid.ProjectId("proj-1"),
		ServiceStackId: uuid.ServiceStackId("svc-1"),
		Source:         enum.AppVersionSourceEnum("CLI"),
		Sequence:       types.Int(7),
		Status:         enum.AppVersionStatusEnum("ACTIVE"),
		Build: &output.AppVersionBuild{
			ServiceStackId:   uuid.NewServiceStackIdNull(uuid.ServiceStackId("build-svc-1")),
			ServiceStackName: types.NewStringNull("build-svc"),
		},
	})

	if got.Id != entity.AppVersionId("app-1") ||
		got.OrgId != entity.OrgId("org-1") ||
		got.ProjectId != entity.ProjectId("proj-1") ||
		got.ServiceId != entity.ServiceId("svc-1") ||
		got.Source != entity.AppVersionSource("CLI") ||
		got.Sequence != 7 ||
		got.Status != entity.AppVersionStatus("ACTIVE") {
		t.Errorf("appVersionFromApiOutput mismatch: %+v", got)
	}
	if got.Build == nil {
		t.Fatal("expected non-nil Build")
	}
	if sid, ok := got.Build.ServiceStackId.Get(); !ok || sid != entity.ServiceId("build-svc-1") {
		t.Errorf("Build.ServiceStackId = %v (ok=%v), want build-svc-1", sid, ok)
	}
	if got.Build.ServiceStackName != "build-svc" {
		t.Errorf("Build.ServiceStackName = %q, want build-svc", got.Build.ServiceStackName)
	}
}

func TestProjectFromApiOutputWithOrg(t *testing.T) {
	org := entity.Org{Id: entity.OrgId("org-1"), Name: "Acme Org"}
	got := projectFromApiOutputWithOrg(org, output.Project{
		Id:       uuid.ProjectId("proj-1"),
		ClientId: uuid.ClientId("org-1"),
		Name:     types.String("proj-name"),
		Mode:     enum.ProjectModeEnum("HA"),
		Status:   enum.ProjectStatusEnum("ACTIVE"),
	})

	want := entity.Project{
		Id:      entity.ProjectId("proj-1"),
		Name:    "proj-name",
		Mode:    entity.ProjectMode("HA"),
		OrgId:   entity.OrgId("org-1"),
		OrgName: "Acme Org",
		Status:  entity.ProjectStatus("ACTIVE"),
	}
	if got != want {
		t.Errorf("projectFromApiOutputWithOrg = %+v, want %+v", got, want)
	}
}
