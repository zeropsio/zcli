package version

import "testing"

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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUpdateAvailable(tc.current, tc.latest); got != tc.want {
				t.Fatalf("isUpdateAvailable(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
			}
		})
	}
}
