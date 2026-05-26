package upgrade

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUpdateAvailable(t *testing.T) {
	cases := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"older current", "v1.0.0", "v1.0.1", true},
		{"older current minor", "v1.0.0", "v1.1.0", true},
		{"newer current", "v1.0.1", "v1.0.0", false},
		{"equal", "v1.0.0", "v1.0.0", false},
		{"local current", "local", "v1.0.0", false},
		{"empty current", "", "v1.0.0", false},
		{"empty latest", "v1.0.0", "", false},
		{"garbage latest", "v1.0.0", "not-a-version", false},
		// git describe stamps commits past a tag as build metadata (see Makefile
		// PROD_VERSION). semver ignores the `+...` suffix in comparisons, so a
		// local build ahead of the released tag must not warn.
		{"build metadata ties with tag", "v1.0.67+11.g03aedf4", "v1.0.67", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equalf(t, tc.want, isUpdateAvailable(tc.current, tc.latest), "isUpdateAvailable(%q, %q)", tc.current, tc.latest)
		})
	}
}

func TestMismatchWarning(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ZEROPS_CLI_DATA_FILE_PATH", filepath.Join(dir, "cli.data"))

	require.NoError(t, writeCacheEntry(&apiResponse{TagName: "v1.2.0"}), "seed cache")

	t.Run("non-semver current returns empty", func(t *testing.T) {
		assert.Empty(t, (Upgrader{current: "local"}).MismatchWarning(), "local current should produce no warning")
	})

	t.Run("equal versions return empty", func(t *testing.T) {
		assert.Empty(t, (Upgrader{current: "v1.2.0"}).MismatchWarning(), "equal versions should produce no warning")
	})

	t.Run("channel hint included", func(t *testing.T) {
		cases := []struct {
			stamp string
			want  string
		}{
			{"npm", "npm install -g @zerops/zcli"},
			{"brew", "brew upgrade zcli"},
			{"nix", "rebuild your profile or flake"},
			{"manual", "zcli upgrade"},
		}
		for _, tc := range cases {
			got := (Upgrader{current: "v1.0.0", channel: tc.stamp}).MismatchWarning()
			assert.Containsf(t, got, "v1.2.0", "channel %q: missing latest version", tc.stamp)
			assert.Containsf(t, got, "v1.0.0", "channel %q: missing current version", tc.stamp)
			assert.Containsf(t, got, tc.want, "channel %q: missing channel hint", tc.stamp)
		}
	})
}
