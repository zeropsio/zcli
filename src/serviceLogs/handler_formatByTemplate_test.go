package serviceLogs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateFixOk(t *testing.T) {
	assert := require.New(t)
	got, err := fixTemplate("{{.timestamp}} {{.severityLabel}} {{.severity  }} {{.facility}} {{ .message}}")
	assert.NoError(err)
	assert.Equal(
		"{{.Timestamp}} {{.SeverityLabel}} {{.Severity}} {{.Facility}} {{.Message}}",
		got,
	)
}

func TestTemplateFixNotOk(t *testing.T) {
	assert := require.New(t)
	_, err := fixTemplate("{{.timestamp}}  {{.severityLabel}} {{.severity  }} {{.facility}} {{ .message}}")
	assert.Error(err)
}
