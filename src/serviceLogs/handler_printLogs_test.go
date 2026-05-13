package serviceLogs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func TestMakeQueryParams_AllFieldsPresent(t *testing.T) {
	got := makeQueryParams(
		InputValues{
			limit:       50,
			minSeverity: 4,
			facility:    16,
			mode:        STREAM,
			tags:        []string{"a", "b"},
		},
		uuid.ServiceStackId("svc-1"),
		uuid.ContainerId("c-1"),
	)
	require.Equal(t,
		"&limit=50&desc=0&facility=16&serviceStackId=svc-1&minimumSeverity=4&tags=a,b&containerId=c-1",
		got,
	)
}

func TestMakeQueryParams_ResponseModeFlipsDesc(t *testing.T) {
	got := makeQueryParams(
		InputValues{limit: 10, minSeverity: -1, facility: 16, mode: RESPONSE},
		uuid.ServiceStackId("svc-1"),
		"",
	)
	require.True(t, strings.Contains(got, "desc=1"), "got: %s", got)
	require.False(t, strings.Contains(got, "minimumSeverity"))
	require.False(t, strings.Contains(got, "tags="))
	require.False(t, strings.Contains(got, "containerId="))
}

func TestMakeQueryParams_OmitsOptionalFields(t *testing.T) {
	cases := []struct {
		name   string
		inputs InputValues
		want   []string
		notin  []string
	}{
		{
			name:   "no severity",
			inputs: InputValues{limit: 10, minSeverity: -1, facility: 16, mode: STREAM, tags: []string{"x"}},
			want:   []string{"tags=x"},
			notin:  []string{"minimumSeverity"},
		},
		{
			name:   "no tags",
			inputs: InputValues{limit: 10, minSeverity: 2, facility: 16, mode: STREAM},
			want:   []string{"minimumSeverity=2"},
			notin:  []string{"tags="},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := makeQueryParams(tc.inputs, uuid.ServiceStackId("svc-1"), "")
			for _, want := range tc.want {
				require.Contains(t, got, want)
			}
			for _, notin := range tc.notin {
				require.NotContains(t, got, notin)
			}
		})
	}
}

func TestGetDesc(t *testing.T) {
	require.Equal(t, 1, getDesc(RESPONSE))
	require.Equal(t, 0, getDesc(STREAM))
	require.Equal(t, 0, getDesc("anything-else"))
}

func TestGetWsUrl(t *testing.T) {
	got := getWsUrl("https://api.example.com/log?token=abc")
	require.Equal(t, "wss://api.example.com/log/stream?token=abc", got)
}

func TestGetWsUrl_PreservesQueryString(t *testing.T) {
	got := getWsUrl("https://api.example.com/log?a=1&b=2")
	require.Equal(t, "wss://api.example.com/log/stream?a=1&b=2", got)
}

func TestGetWsUrl_NoQueryString(t *testing.T) {
	// Prior implementation panicked on URLs without "?". New behavior:
	// wss-rewrite and append /stream without a trailing "?".
	got := getWsUrl("https://api.example.com/log")
	require.Equal(t, "wss://api.example.com/log/stream", got)
}
