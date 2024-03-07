package serviceLogs

import (
	"testing"
)

func TestTemplateFixOk(t *testing.T) {
	got, _ := fixTemplate("{{.timestamp}} {{.severityLabel}} {{.severity  }} {{.facility}} {{ .message}}")
	want := "{{.Timestamp}} {{.SeverityLabel}} {{.Severity}} {{.Facility}} {{.Message}}"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestTemplateFixNotOk(t *testing.T) {
	_, err := fixTemplate("{{.timestamp}}  {{.severityLabel}} {{.severity  }} {{.facility}} {{ .message}}")

	if err == nil {
		t.Errorf("got %v, wanted %q", nil, err)
	}
}
