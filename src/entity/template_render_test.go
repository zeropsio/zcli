package entity

import (
	"strings"
	"testing"
	"text/template"
)

// Pins what `--out` templates render for the fields commands actually expose, so
// the domain-type swap (zerops-go wrappers -> native/typed-zcli types) can't change
// the strings users see. Field names must stay identical too - renaming any of
// these would break existing user templates.
func render(t *testing.T, tmpl string, data any) string {
	t.Helper()
	parsed, err := template.New("out").Parse(tmpl)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var sb strings.Builder
	if err := parsed.Execute(&sb, data); err != nil {
		t.Fatalf("execute: %v", err)
	}
	return sb.String()
}

func TestServiceTemplateRender(t *testing.T) {
	svc := Service{
		Id:                          ServiceId("svc-123"),
		ProjectId:                   ProjectId("proj-456"),
		OrgId:                       OrgId("org-789"),
		Name:                        "api",
		Status:                      ServiceStatus("ACTIVE"),
		ServiceTypeId:               ServiceTypeId("nodejs@20"),
		ServiceStackTypeVersionName: "20",
	}

	got := render(t, "{{.Id}}|{{.ProjectId}}|{{.OrgId}}|{{.Name}}|{{.Status}}|{{.ServiceTypeId}}|{{.ServiceStackTypeVersionName}}", svc)
	want := "svc-123|proj-456|org-789|api|ACTIVE|nodejs@20|20"
	if got != want {
		t.Errorf("service render = %q, want %q", got, want)
	}
}

func TestProjectTemplateRender(t *testing.T) {
	proj := Project{
		Id:          ProjectId("proj-456"),
		Name:        "my-project",
		Mode:        ProjectMode("HA"),
		OrgId:       OrgId("org-789"),
		OrgName:     "my-org",
		Description: "desc",
		Status:      ProjectStatus("ACTIVE"),
	}

	got := render(t, "{{.Id}}|{{.Name}}|{{.Mode}}|{{.OrgId}}|{{.OrgName}}|{{.Description}}|{{.Status}}", proj)
	want := "proj-456|my-project|HA|org-789|my-org|desc|ACTIVE"
	if got != want {
		t.Errorf("project render = %q, want %q", got, want)
	}
}

// Guards that the exposed template field-name lists stay stable.
func TestEntityFieldNamesStable(t *testing.T) {
	wantService := []string{
		"{{.Id}}", "{{.ProjectId}}", "{{.OrgId}}", "{{.Name}}", "{{.Status}}",
		"{{.ServiceTypeId}}", "{{.ServiceTypeCategory}}", "{{.ServiceStackTypeVersionName}}",
	}
	if strings.Join(ServiceFields, ",") != strings.Join(wantService, ",") {
		t.Errorf("ServiceFields = %v, want %v", ServiceFields, wantService)
	}

	wantProject := []string{
		"{{.Id}}", "{{.Name}}", "{{.Mode}}", "{{.OrgId}}", "{{.OrgName}}",
		"{{.Description}}", "{{.Status}}",
	}
	if strings.Join(ProjectFields, ",") != strings.Join(wantProject, ",") {
		t.Errorf("ProjectFields = %v, want %v", ProjectFields, wantProject)
	}
}
