package serviceLogs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateUri_WithoutLastMsgId(t *testing.T) {
	h := &Handler{}
	got := h.updateUri("wss://api/stream", "?limit=10")
	require.Equal(t, "wss://api/stream?limit=10", got)
}

func TestUpdateUri_AppendsFromOnReconnect(t *testing.T) {
	h := &Handler{lastMsgId: "abc"}
	got := h.updateUri("wss://api/stream", "?limit=10")
	require.Equal(t, "wss://api/stream?limit=10&from=abc", got)
}

func TestPrintStreamLog_UpdatesLastMsgIdToFinalItem(t *testing.T) {
	h := &Handler{out: &discardWriter{}}
	body, err := json.Marshal(Response{Items: []Data{
		{Id: "first", Timestamp: "2024-01-02T10:00:00.000000Z", Message: "a"},
		{Id: "last", Timestamp: "2024-01-02T10:00:01.000000Z", Message: "b"},
	}})
	require.NoError(t, err)

	require.NoError(t, h.printStreamLog(body, FULL, "{{.Message}}"))
	require.Equal(t, "last", h.lastMsgId)
}

func TestPrintStreamLog_EmptyItemsLeavesIdUntouched(t *testing.T) {
	h := &Handler{out: &discardWriter{}, lastMsgId: "previous"}
	body, err := json.Marshal(Response{Items: nil})
	require.NoError(t, err)

	require.NoError(t, h.printStreamLog(body, FULL, "{{.Message}}"))
	require.Equal(t, "previous", h.lastMsgId)
}

func TestPrintStreamLog_MalformedJSONIsCurrentlySilent(t *testing.T) {
	// Current behavior: parseResponse error is discarded, function returns
	// nil, and lastMsgId is left unchanged. Pinning this so the refactor
	// item that proposes propagating the JSON error has a failing test to
	// flip.
	h := &Handler{out: &discardWriter{}, lastMsgId: "previous"}
	require.NoError(t, h.printStreamLog([]byte("not json"), FULL, "{{.Message}}"))
	require.Equal(t, "previous", h.lastMsgId)
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
