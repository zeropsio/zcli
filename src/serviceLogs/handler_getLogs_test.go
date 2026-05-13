package serviceLogs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseResponse_Valid(t *testing.T) {
	body, err := json.Marshal(Response{Items: []Data{{Id: "x", Message: "hello"}}})
	require.NoError(t, err)
	got, err := parseResponse(body)
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	require.Equal(t, "x", got.Items[0].Id)
	require.Equal(t, "hello", got.Items[0].Message)
}

func TestParseResponse_EmptyItems(t *testing.T) {
	got, err := parseResponse([]byte(`{"items":[]}`))
	require.NoError(t, err)
	require.Empty(t, got.Items)
}

func TestParseResponse_InvalidJSONReturnsError(t *testing.T) {
	_, err := parseResponse([]byte(`not json`))
	require.Error(t, err)
}

func TestGetLogs_HappyPath(t *testing.T) {
	// Drive getLogs through an httptest server: it should fetch, decode, and
	// emit the formatted output via parseResponseByFormat.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_ = json.NewEncoder(w).Encode(Response{Items: []Data{
			{Id: "1", Timestamp: "2024-01-02T10:00:00.000000Z", Message: "hello"},
		}})
	}))
	defer srv.Close()

	var buf strings.Builder
	h := &Handler{out: &buf}
	require.NoError(t, h.getLogs(context.Background(), http.MethodGet, srv.URL, FULL, "{{.Message}}", STREAM))
	require.Equal(t, "hello\n", buf.String())
}

func TestGetLogs_NonJSONBodyReturnsError(t *testing.T) {
	// Locks in current behavior: any non-JSON body (including an HTTP error
	// page from a 4xx/5xx) surfaces as a JSON unmarshal error rather than a
	// "request failed" message. The review-noted refactor would add a status
	// check before parsing — update this test then.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("upstream broke"))
	}))
	defer srv.Close()

	h := &Handler{out: &discardWriter{}}
	err := h.getLogs(context.Background(), http.MethodGet, srv.URL, FULL, "", STREAM)
	require.Error(t, err)
}

func TestGetLogs_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := &Handler{out: &discardWriter{}}
	err := h.getLogs(ctx, http.MethodGet, srv.URL, FULL, "", STREAM)
	require.Error(t, err)
}
