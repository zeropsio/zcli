package serviceLogs

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestTemplateFixOk(t *testing.T) {
	got, err := fixTemplate("{{.timestamp}} {{.severityLabel}} {{.severity  }} {{.facility}} {{ .message}}")
	require.NoError(t, err)
	require.Equal(t,
		"{{.Timestamp}} {{.SeverityLabel}} {{.Severity}} {{.Facility}} {{.Message}}",
		got,
	)
}

func TestTemplateFix_AlreadyCapitalized(t *testing.T) {
	got, err := fixTemplate("{{.Timestamp}} {{.Message}}")
	require.NoError(t, err)
	require.Equal(t, "{{.Timestamp}} {{.Message}}", got)
}

func TestTemplateFix_DoubleSpaceRejected(t *testing.T) {
	// fixTemplate first splits on the literal "} {" separator, so any input
	// where tokens aren't separated by exactly one space leaves a stray "{"
	// in a token and testTokens rejects it.
	_, err := fixTemplate("{{.timestamp}}  {{.severityLabel}} {{.severity  }} {{.facility}} {{ .message}}")
	require.Error(t, err)
}

func TestTemplateFix_SingleToken(t *testing.T) {
	got, err := fixTemplate("{{.message}}")
	require.NoError(t, err)
	require.Equal(t, "{{.Message}}", got)
}

func TestFormatDataByTemplate_Renders(t *testing.T) {
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	tmpl, err := template.New("").Parse("{{.Hostname}}: {{.Message}}")
	require.NoError(t, err)
	require.NoError(t, h.formatDataByTemplate(
		Data{Message: "hello", Hostname: "host-a"},
		tmpl,
	))
	require.Equal(t, "host-a: hello\n", buf.String())
}

func TestGetFullWithTemplate_NormalizesAndRendersAllItems(t *testing.T) {
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	require.NoError(t, h.getFullWithTemplate([]Data{
		{Message: "first"},
		{Message: "second"},
	}, "{{.message}}"))
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Equal(t, []string{"first", "second"}, lines)
}

func TestGetFullWithTemplate_PropagatesBadTemplate(t *testing.T) {
	h := &Handler{out: &bytes.Buffer{}}
	err := h.getFullWithTemplate([]Data{{Message: "x"}}, "{{.timestamp}}  {{.message}}")
	require.Error(t, err)
}
