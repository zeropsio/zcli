package browser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testURL = "https://app.zerops.io/device?code=ABCD-EFGH&lang=en"

func TestCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goos      string
		env       map[string]string
		osRelease string
		onPath    []string
		want      []string
	}{
		{
			name: "darwin",
			goos: "darwin",
			want: []string{"open", testURL},
		},
		{
			name: "windows",
			goos: "windows",
			want: []string{"rundll32", "url.dll,FileProtocolHandler", testURL},
		},
		{
			name:   "wsl via env",
			goos:   "linux",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu"},
			onPath: []string{"wslview", "rundll32.exe", "xdg-open"},
			want:   []string{"wslview", testURL},
		},
		{
			name:      "wsl via osrelease",
			goos:      "linux",
			osRelease: "6.6.87.2-microsoft-standard-WSL2\n",
			onPath:    []string{"wslview", "explorer.exe", "xdg-open"},
			want:      []string{"wslview", testURL},
		},
		{
			name:      "wsl via osrelease in upper case",
			goos:      "linux",
			osRelease: "4.4.0-19041-Microsoft\n",
			onPath:    []string{"rundll32.exe"},
			want:      []string{"rundll32.exe", "url.dll,FileProtocolHandler", testURL},
		},
		{
			name:   "wsl without wslview uses rundll32.exe",
			goos:   "linux",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu"},
			onPath: []string{"rundll32.exe", "explorer.exe", "xdg-open"},
			want:   []string{"rundll32.exe", "url.dll,FileProtocolHandler", testURL},
		},
		{
			name:   "wsl without windows path uses absolute rundll32.exe",
			goos:   "linux",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu", "DISPLAY": ":0"},
			onPath: []string{"/mnt/c/Windows/System32/rundll32.exe", "xdg-open"},
			want:   []string{"/mnt/c/Windows/System32/rundll32.exe", "url.dll,FileProtocolHandler", testURL},
		},
		{
			name:   "wsl without windows launchers falls back to xdg-open",
			goos:   "linux",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu", "DISPLAY": ":0"},
			onPath: []string{"explorer.exe", "xdg-open"},
			want:   []string{"xdg-open", testURL},
		},
		{
			name: "wsl without windows launchers and display",
			goos: "linux",
			env:  map[string]string{"WSL_DISTRO_NAME": "Ubuntu"},
		},
		{
			name:   "wsl with display picks wslview",
			goos:   "linux",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu", "DISPLAY": ":0", "WAYLAND_DISPLAY": "wayland-0"},
			onPath: []string{"wslview", "rundll32.exe", "xdg-open"},
			want:   []string{"wslview", testURL},
		},
		{
			name:   "wsl with display picks rundll32.exe",
			goos:   "linux",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu", "DISPLAY": ":0"},
			onPath: []string{"rundll32.exe", "xdg-open"},
			want:   []string{"rundll32.exe", "url.dll,FileProtocolHandler", testURL},
		},
		{
			name:      "linux with display",
			goos:      "linux",
			env:       map[string]string{"DISPLAY": ":0"},
			osRelease: "6.8.0-45-generic\n",
			onPath:    []string{"xdg-open", "wslview"},
			want:      []string{"xdg-open", testURL},
		},
		{
			name:   "linux with wayland display",
			goos:   "linux",
			env:    map[string]string{"WAYLAND_DISPLAY": "wayland-0"},
			onPath: []string{"xdg-open"},
			want:   []string{"xdg-open", testURL},
		},
		{
			name:   "linux headless",
			goos:   "linux",
			onPath: []string{"xdg-open", "www-browser", "x-www-browser"},
		},
		{
			name: "linux headless ignores BROWSER",
			goos: "linux",
			env:  map[string]string{"BROWSER": "lynx"},
		},
		{
			name:   "linux without xdg-open",
			goos:   "linux",
			env:    map[string]string{"DISPLAY": ":0", "BROWSER": "firefox"},
			onPath: []string{"www-browser", "x-www-browser"},
		},
		{
			name:   "freebsd with display",
			goos:   "freebsd",
			env:    map[string]string{"DISPLAY": ":0"},
			onPath: []string{"xdg-open"},
			want:   []string{"xdg-open", testURL},
		},
		{
			name:   "freebsd ignores wsl env",
			goos:   "freebsd",
			env:    map[string]string{"WSL_DISTRO_NAME": "Ubuntu", "DISPLAY": ":0"},
			onPath: []string{"xdg-open", "wslview"},
			want:   []string{"xdg-open", testURL},
		},
		{
			name: "unknown goos",
			goos: "plan9",
			env:  map[string]string{"DISPLAY": ":0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := host{
				goos:   tt.goos,
				getenv: func(key string) string { return tt.env[key] },
				readFile: func(name string) ([]byte, error) {
					if name != "/proc/sys/kernel/osrelease" || tt.osRelease == "" {
						return nil, os.ErrNotExist
					}
					return []byte(tt.osRelease), nil
				},
				lookPath: func(file string) (string, error) {
					for _, p := range tt.onPath {
						if p == file {
							return "/usr/bin/" + file, nil
						}
					}
					return "", os.ErrNotExist
				},
			}

			got, err := h.command(testURL)
			if tt.want == nil {
				require.ErrorIs(t, err, ErrNoBrowser)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestOpenRejectsNonWebURL(t *testing.T) {
	t.Parallel()

	for _, rawURL := range []string{"file:///etc/passwd", "C:\\Windows\\System32\\calc.exe", "-a Calculator", "javascript:alert(1)"} {
		require.Error(t, Open(rawURL), rawURL)
	}
}
