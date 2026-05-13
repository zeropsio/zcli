package serviceLogs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func sampleLogs() []Data {
	return []Data{
		{
			Timestamp:     "2024-01-02T10:00:00.000000Z",
			Hostname:      "host-a",
			Priority:      14,
			Tag:           "tag-a",
			AppName:       "app",
			ProcId:        "1",
			MsgId:         "m1",
			Message:       "first",
			Content:       "first-content",
			SeverityLabel: "INFO",
			Severity:      6,
			Id:            "id-1",
		},
		{
			Timestamp:     "2024-01-02T10:00:01.000000Z",
			Hostname:      "host-b",
			Priority:      14,
			Tag:           "tag-b",
			AppName:       "app",
			ProcId:        "2",
			MsgId:         "m2",
			Message:       "second",
			Content:       "second-content",
			SeverityLabel: "INFO",
			Severity:      6,
			Id:            "id-2",
		},
	}
}

func runFormat(t *testing.T, items []Data, format, formatTemplate, mode string) string {
	t.Helper()
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	require.NoError(t, h.parseResponseByFormat(Response{Items: items}, format, formatTemplate, mode))
	return buf.String()
}

func TestParseResponseByFormat_FULL_DefaultsToRFC5424(t *testing.T) {
	out := runFormat(t, sampleLogs(), FULL, "", STREAM)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 2)
	// <prio>1 timestamp hostname appName procId msgId - message
	require.Equal(t, "<14>1 2024-01-02T10:00:00.000000Z host-a app 1 m1 - first", lines[0])
	require.Equal(t, "<14>1 2024-01-02T10:00:01.000000Z host-b app 2 m2 - second", lines[1])
}

func TestParseResponseByFormat_FULL_WithTemplate(t *testing.T) {
	out := runFormat(t, sampleLogs(), FULL, "{{.Message}}", STREAM)
	require.Equal(t, "first\nsecond\n", out)
}

func TestParseResponseByFormat_SHORT(t *testing.T) {
	out := runFormat(t, sampleLogs(), SHORT, "", STREAM)
	require.Equal(t,
		"2024-01-02T10:00:00.000000Z first-content \n2024-01-02T10:00:01.000000Z second-content \n",
		out,
	)
}

func TestParseResponseByFormat_JSONSTREAM_OneObjectPerLine(t *testing.T) {
	out := runFormat(t, sampleLogs(), JSONSTREAM, "", STREAM)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 2)
	var one Data
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &one))
	require.Equal(t, "first", one.Message)
}

func TestParseResponseByFormat_JSON_ArrayDefault(t *testing.T) {
	// Anything that isn't FULL/SHORT/JSONSTREAM falls into the JSON-array
	// default branch (including the literal "JSON" string).
	out := runFormat(t, sampleLogs(), JSON, "", STREAM)
	var arr []Data
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &arr))
	require.Len(t, arr, 2)
	require.Equal(t, "first", arr[0].Message)
}

func TestParseResponseByFormat_ResponseModeReverses(t *testing.T) {
	// RESPONSE mode is the synchronous "last N logs" request: items arrive
	// newest-first, and parseResponseByFormat reverses them so output is
	// oldest-first.
	logs := sampleLogs()
	out := runFormat(t, logs, SHORT, "", RESPONSE)
	require.True(t, strings.Index(out, "second-content") < strings.Index(out, "first-content"),
		"RESPONSE mode should print items in reverse order, got:\n%s", out)
}

func TestParseResponseByFormat_StreamModeKeepsOrder(t *testing.T) {
	logs := sampleLogs()
	out := runFormat(t, logs, SHORT, "", STREAM)
	require.True(t, strings.Index(out, "first-content") < strings.Index(out, "second-content"))
}

func TestReverseLogs(t *testing.T) {
	cases := []struct {
		name string
		in   []Data
		want []string
	}{
		{"empty", nil, nil},
		{"single", []Data{{Id: "a"}}, []string{"a"}},
		{"even", []Data{{Id: "a"}, {Id: "b"}}, []string{"b", "a"}},
		{"odd", []Data{{Id: "a"}, {Id: "b"}, {Id: "c"}}, []string{"c", "b", "a"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := reverseLogs(tc.in)
			ids := make([]string, len(got))
			for i, d := range got {
				ids[i] = d.Id
			}
			if tc.want == nil {
				require.Empty(t, ids)
				return
			}
			require.Equal(t, tc.want, ids)
		})
	}
}

func TestReverseLogs_MutatesInputSlice(t *testing.T) {
	// reverseLogs reverses in place. Lock that in — callers in handler_formatLogs.go
	// pass the request items directly, expecting the same backing array.
	in := []Data{{Id: "a"}, {Id: "b"}, {Id: "c"}}
	got := reverseLogs(in)
	require.Equal(t, "c", in[0].Id, "input was not mutated; reverseLogs is supposed to reverse in place")
	require.Equal(t, "c", got[0].Id)
}
