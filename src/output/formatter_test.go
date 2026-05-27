package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"", FormatTable, false},
		{"table", FormatTable, false},
		{"json", FormatJSON, false},
		{"csv", FormatCSV, false},
		{"JSON", FormatJSON, false},
		{"CSV", FormatCSV, false},
		{"yaml", FormatTable, true},
		{"xml", FormatTable, true},
	}
	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input %q", tt.input)
		} else {
			require.NoError(t, err, "input %q", tt.input)
			require.Equal(t, tt.want, got, "input %q", tt.input)
		}
	}
}

func TestPrintDataTable(t *testing.T) {
	headers := []string{"id", "name", "status"}
	rows := [][]string{
		{"1", "foo", "active"},
		{"2", "bar", "inactive"},
	}
	result, err := PrintData(headers, rows, FormatTable)
	require.NoError(t, err)
	require.Contains(t, result, "ID")
	require.Contains(t, result, "foo")
	require.Contains(t, result, "bar")
}

func TestPrintDataJSON(t *testing.T) {
	headers := []string{"id", "name", "status"}
	rows := [][]string{
		{"1", "foo", "active"},
		{"2", "bar", "inactive"},
	}
	result, err := PrintData(headers, rows, FormatJSON)
	require.NoError(t, err)

	require.Contains(t, result, `"id": "1"`)
	require.Contains(t, result, `"name": "foo"`)
	require.Contains(t, result, `"status": "active"`)
	require.Contains(t, result, `"id": "2"`)
	require.Contains(t, result, `"name": "bar"`)

	var parsed []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &parsed))
	require.Len(t, parsed, 2)
}

func TestPrintDataCSV(t *testing.T) {
	headers := []string{"id", "name", "status"}
	rows := [][]string{
		{"1", "foo", "active"},
		{"2", "bar", "inactive"},
	}
	result, err := PrintData(headers, rows, FormatCSV)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	require.Len(t, lines, 3)
	require.Equal(t, "id,name,status", lines[0])
	require.Equal(t, "1,foo,active", lines[1])
	require.Equal(t, "2,bar,inactive", lines[2])
}

func TestPrintDataEmpty(t *testing.T) {
	headers := []string{"id", "name"}
	var rows [][]string

	result, err := PrintData(headers, rows, FormatJSON)
	require.NoError(t, err)
	require.Equal(t, "[]", result)

	result, err = PrintData(headers, rows, FormatCSV)
	require.NoError(t, err)
	require.Equal(t, "id,name\n", result)

	result, err = PrintData(headers, rows, FormatTable)
	require.NoError(t, err)
	require.Contains(t, result, "ID")
	require.Contains(t, result, "NAME")
}
