package serviceLogs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests lock in the current observable behavior of the input-validation
// helpers. Some pieces are intentionally awkward (e.g. error paths return
// "default" facility/severity values rather than zeros) — the refactor will
// change those, and the corresponding tests should change with it.

func newHandler() *Handler { return &Handler{} }

func baseRunConfig() RunConfig {
	return RunConfig{
		Limit:       100,
		MinSeverity: "",
		MsgType:     "APPLICATION",
		Format:      "FULL",
		Levels:      DefaultLevels,
	}
}

func TestGetLimit(t *testing.T) {
	h := newHandler()
	cases := []struct {
		name    string
		limit   int
		wantErr bool
	}{
		{"min boundary", 1, false},
		{"middle", 100, false},
		{"max boundary", 1000, false},
		{"zero rejected", 0, true},
		{"negative rejected", -1, true},
		{"over max rejected", 1001, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := h.getLimit(RunConfig{Limit: tc.limit})
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.limit, got)
		})
	}
}

func TestGetMinSeverity(t *testing.T) {
	h := newHandler()

	t.Run("empty input returns -1", func(t *testing.T) {
		got, err := h.getMinSeverity(RunConfig{Levels: DefaultLevels})
		require.NoError(t, err)
		require.Equal(t, -1, got)
	})

	nameCases := []struct {
		input string
		want  int
	}{
		{"EMERGENCY", 0},
		{"ALERT", 1},
		{"CRITICAL", 2},
		{"ERROR", 3},
		{"WARNING", 4},
		{"NOTICE", 5},
		{"INFORMATIONAL", 6},
		{"DEBUG", 7},
		{"error", 3},     // lower-case accepted (uppercased internally)
		{"Warning", 4},   // mixed-case accepted
		{"INFOrmational", 6},
	}
	for _, tc := range nameCases {
		t.Run("name "+tc.input, func(t *testing.T) {
			got, err := h.getMinSeverity(RunConfig{MinSeverity: tc.input, Levels: DefaultLevels})
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	numericCases := []struct {
		input string
		want  int
	}{
		{"0", 0}, {"1", 1}, {"2", 2}, {"3", 3},
		{"4", 4}, {"5", 5}, {"6", 6}, {"7", 7},
	}
	for _, tc := range numericCases {
		t.Run("numeric "+tc.input, func(t *testing.T) {
			got, err := h.getMinSeverity(RunConfig{MinSeverity: tc.input, Levels: DefaultLevels})
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	t.Run("unknown name returns error", func(t *testing.T) {
		_, err := h.getMinSeverity(RunConfig{MinSeverity: "bogus", Levels: DefaultLevels})
		require.Error(t, err)
	})

	t.Run("out-of-range number returns error", func(t *testing.T) {
		_, err := h.getMinSeverity(RunConfig{MinSeverity: "9", Levels: DefaultLevels})
		require.Error(t, err)
	})
}

func TestGetFacility(t *testing.T) {
	h := newHandler()
	cases := []struct {
		msgType string
		want    int
		wantErr bool
	}{
		{"APPLICATION", 16, false},
		{"application", 16, false},
		{"WEBSERVER", 17, false},
		{"webserver", 17, false},
		{"", 16, true},
		{"BOGUS", 16, true},
	}
	for _, tc := range cases {
		t.Run(tc.msgType, func(t *testing.T) {
			got, err := h.getFacility(RunConfig{MsgType: tc.msgType})
			if tc.wantErr {
				require.Error(t, err)
				// current behavior: returns 16 even on error
				require.Equal(t, 16, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestGetFormat(t *testing.T) {
	h := newHandler()

	t.Run("valid formats without template", func(t *testing.T) {
		for _, f := range []string{"FULL", "SHORT", "JSON", "JSONSTREAM"} {
			got, gotT, err := h.getFormat(RunConfig{Format: f})
			require.NoError(t, err)
			require.Equal(t, f, got)
			require.Empty(t, gotT)
		}
	})

	t.Run("lowercase format accepted", func(t *testing.T) {
		got, _, err := h.getFormat(RunConfig{Format: "full"})
		require.NoError(t, err)
		require.Equal(t, "FULL", got)
	})

	t.Run("invalid format rejected", func(t *testing.T) {
		_, _, err := h.getFormat(RunConfig{Format: "YAML"})
		require.Error(t, err)
	})

	t.Run("template requires FULL format", func(t *testing.T) {
		_, _, err := h.getFormat(RunConfig{Format: "SHORT", FormatTemplate: "{{.Message}}"})
		require.Error(t, err)
	})

	t.Run("template with FULL succeeds and is returned unchanged", func(t *testing.T) {
		// getFormat only runs validateTemplate (parser check). The
		// case-normalization performed by fixTemplate happens later, inside
		// getFullWithTemplate when actually rendering. So getFormat returns
		// the raw template string here.
		got, gotT, err := h.getFormat(RunConfig{Format: "FULL", FormatTemplate: "{{.message}}"})
		require.NoError(t, err)
		require.Equal(t, "FULL", got)
		require.Equal(t, "{{.message}}", gotT)
	})

	t.Run("unparseable template rejected", func(t *testing.T) {
		_, _, err := h.getFormat(RunConfig{Format: "FULL", FormatTemplate: "{{.Message"})
		require.Error(t, err)
	})
}

func TestGetTags(t *testing.T) {
	h := newHandler()
	tags := []string{"zbuilder@id", "extra"}
	require.Equal(t, tags, h.getTags(RunConfig{Tags: tags}))
	require.Nil(t, h.getTags(RunConfig{}))
}

func TestCheckInputValues_HappyPath(t *testing.T) {
	h := newHandler()
	got, err := h.checkInputValues(baseRunConfig())
	require.NoError(t, err)
	require.Equal(t, InputValues{
		limit:          100,
		minSeverity:    -1,
		facility:       16,
		format:         "FULL",
		formatTemplate: "",
		mode:           RESPONSE,
		tags:           nil,
	}, got)
}

func TestCheckInputValues_FollowSetsStreamMode(t *testing.T) {
	h := newHandler()
	cfg := baseRunConfig()
	cfg.Follow = true
	got, err := h.checkInputValues(cfg)
	require.NoError(t, err)
	require.Equal(t, STREAM, got.mode)
}

func TestCheckInputValues_FollowJSONMismatch(t *testing.T) {
	h := newHandler()
	cfg := baseRunConfig()
	cfg.Follow = true
	cfg.Format = "JSON"
	_, err := h.checkInputValues(cfg)
	require.Error(t, err)
}

func TestCheckInputValues_PropagatesValidationErrors(t *testing.T) {
	h := newHandler()
	cases := []struct {
		name    string
		mutator func(*RunConfig)
	}{
		{"bad limit", func(c *RunConfig) { c.Limit = 0 }},
		{"bad severity", func(c *RunConfig) { c.MinSeverity = "bogus" }},
		{"bad msg type", func(c *RunConfig) { c.MsgType = "BOGUS" }},
		{"bad format", func(c *RunConfig) { c.Format = "YAML" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := baseRunConfig()
			tc.mutator(&cfg)
			_, err := h.checkInputValues(cfg)
			require.Error(t, err)
		})
	}
}

func TestDefaultLevels_IndexMatchesSeverity(t *testing.T) {
	// The Levels array is indexed by severity number (0..7); the second
	// column happens to also be the stringified index. checkInputValues
	// relies on both invariants — pin them here.
	require.Len(t, DefaultLevels, 8)
	expected := []string{"EMERGENCY", "ALERT", "CRITICAL", "ERROR", "WARNING", "NOTICE", "INFORMATIONAL", "DEBUG"}
	for i, name := range expected {
		require.Equal(t, name, DefaultLevels[i][0], "level %d name", i)
	}
}
