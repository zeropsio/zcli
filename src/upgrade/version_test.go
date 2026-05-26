package upgrade

import (
	"path/filepath"
	"strings"
	"testing"
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
			if got := isUpdateAvailable(tc.current, tc.latest); got != tc.want {
				t.Fatalf("isUpdateAvailable(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
			}
		})
	}
}

func TestMismatchWarning(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ZEROPS_CLI_DATA_FILE_PATH", filepath.Join(dir, "cli.data"))

	if err := writeCacheEntry(&apiResponse{TagName: "v1.2.0"}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	t.Run("non-semver current returns empty", func(t *testing.T) {
		u := Upgrader{current: "local"}
		if got := u.MismatchWarning(); got != "" {
			t.Errorf("local: want empty, got %q", got)
		}
	})

	t.Run("equal versions return empty", func(t *testing.T) {
		u := Upgrader{current: "v1.2.0"}
		if got := u.MismatchWarning(); got != "" {
			t.Errorf("equal: want empty, got %q", got)
		}
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
			u := Upgrader{current: "v1.0.0", channel: tc.stamp}
			got := u.MismatchWarning()
			if !strings.Contains(got, "v1.2.0") || !strings.Contains(got, "v1.0.0") {
				t.Errorf("channel %q: %q missing version info", tc.stamp, got)
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("channel %q: %q missing hint %q", tc.stamp, got, tc.want)
			}
		}
	})
}
