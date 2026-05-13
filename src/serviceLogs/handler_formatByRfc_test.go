package serviceLogs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixTimestamp(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "already 27 chars unchanged",
			in:   "2024-01-02T10:00:00.000000Z",
			want: "2024-01-02T10:00:00.000000Z",
		},
		{
			name: "shorter millis padded to 6",
			in:   "2024-01-02T10:00:00.123Z",
			want: "2024-01-02T10:00:00.123000Z",
		},
		{
			// fixTimestamp splits on "." — when the input has no decimal
			// point, splitVal[0] retains the trailing "Z" and the rebuilt
			// string ends up doubly-Z'd. This is current behavior; the
			// refactor item flagged in the review proposes replacing the
			// hand-rolled padding with time.Parse/Format, which would
			// produce "2024-01-02T10:00:00.000000Z". Update this expectation
			// when that lands.
			name: "no millis: known-buggy double-Z output",
			in:   "2024-01-02T10:00:00Z",
			want: "2024-01-02T10:00:00Z.000000Z",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, fixTimestamp(tc.in))
		})
	}
}

func TestRfc3164TimeFormat(t *testing.T) {
	require.Equal(t, "Jan 02 10:00:00", rfc3164TimeFormat("2024-01-02T10:00:00.000000Z"))
}

func TestRfc3164TimeFormat_ReturnsErrorTextOnBadInput(t *testing.T) {
	// Current contract: malformed input is silently surfaced as the raw
	// parse-error string. Pin it so a refactor doesn't quietly change to
	// "" or panic.
	out := rfc3164TimeFormat("not a timestamp")
	require.NotEmpty(t, out)
	require.NotEqual(t, "Jan 02 10:00:00", out)
}

func TestGetVal(t *testing.T) {
	require.Equal(t, "-", getVal(""))
	require.Equal(t, "hello", getVal("hello"))
}

func TestGetFullByRfc_5424(t *testing.T) {
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	require.NoError(t, h.getFullByRfc([]Data{
		{
			Timestamp: "2024-01-02T10:00:00.000000Z",
			Hostname:  "host-a",
			Priority:  14,
			AppName:   "app",
			ProcId:    "p1",
			MsgId:     "m1",
			Message:   "msg",
		},
	}, RFC5424))
	require.Equal(t,
		"<14>1 2024-01-02T10:00:00.000000Z host-a app p1 m1 - msg\n",
		buf.String(),
	)
}

func TestGetFullByRfc_5424_EmptyFieldsBecomeDash(t *testing.T) {
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	require.NoError(t, h.getFullByRfc([]Data{
		{
			Timestamp: "2024-01-02T10:00:00.000000Z",
			Hostname:  "host-a",
			Priority:  14,
			// AppName/ProcId/MsgId empty → rendered as "-"
			Message: "msg",
		},
	}, RFC5424))
	require.Contains(t, buf.String(), "host-a - - - - msg")
}

func TestGetFullByRfc_3164(t *testing.T) {
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	require.NoError(t, h.getFullByRfc([]Data{
		{
			Timestamp: "2024-01-02T10:00:00.000000Z",
			Hostname:  "host-a",
			Priority:  34,
			Tag:       "su",
			Message:   "failed",
		},
	}, RFC3164))
	got := buf.String()
	require.True(t, strings.HasPrefix(got, "<34>"), "got: %q", got)
	require.Contains(t, got, "Jan 02 10:00:00 host-a su: failed")
	require.True(t, strings.HasSuffix(got, "\n"))
}

func TestGetFullByRfc_UnknownRfcUses5424(t *testing.T) {
	// `getFullByRfc` only special-cases RFC3164; everything else falls into
	// the RFC5424 branch. Locking that in so a refactor that switches on the
	// constant doesn't accidentally drop the default.
	var buf bytes.Buffer
	h := &Handler{out: &buf}
	require.NoError(t, h.getFullByRfc([]Data{
		{Timestamp: "2024-01-02T10:00:00.000000Z", Hostname: "h", Priority: 1, Message: "m"},
	}, "unknown"))
	require.Contains(t, buf.String(), ">1 ")
}
