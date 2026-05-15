package version

import (
	"os"
	"path/filepath"
	"strings"
)

type InstallMethod int

const (
	InstallManual InstallMethod = iota
	InstallNix
	InstallNpm
	InstallBrew
)

func (m InstallMethod) String() string {
	switch m {
	case InstallNix:
		return "nix"
	case InstallNpm:
		return "npm"
	case InstallBrew:
		return "homebrew"
	default:
		return "manual"
	}
}

// Hint returns the one-line upgrade instruction for this install channel.
// Always non-empty so warnings always tell the user what to do.
func (m InstallMethod) Hint() string {
	switch m {
	case InstallNix:
		return "Update via Nix: rebuild your profile or flake."
	case InstallNpm:
		return "Update via npm: npm install -g @zerops/zcli"
	case InstallBrew:
		return "Update via Homebrew: brew upgrade zcli"
	default:
		return "Run: zcli upgrade"
	}
}

func (m InstallMethod) IsPackageManager() bool {
	return m != InstallManual
}

// channel is stamped by packagers via `-ldflags "-X .../version.channel=<name>"`.
// Empty means "not stamped" — Detect() falls back to path-based heuristics.
// Packagers who install to paths indistinguishable from a manual install
// (notably AUR and winget MSI installers) should set this.
var channel = ""

// Detect returns the channel the running binary was installed through. The
// build-time channel stamp takes precedence; otherwise we infer from the
// binary path. Falls back to InstallManual when neither yields a result.
func Detect() InstallMethod {
	if m, ok := parseChannel(channel); ok {
		return m
	}
	exe, err := os.Executable()
	if err != nil {
		return InstallManual
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return detectFromPath(exe)
}

// parseChannel maps a build-stamped channel name to an InstallMethod.
// Unknown or empty names return (_, false) so callers fall back to path
// detection rather than silently mis-reporting.
func parseChannel(c string) (InstallMethod, bool) {
	switch c {
	case "manual":
		return InstallManual, true
	case "nix":
		return InstallNix, true
	case "npm":
		return InstallNpm, true
	case "brew":
		return InstallBrew, true
	default:
		return 0, false
	}
}

func detectFromPath(path string) InstallMethod {
	p := filepath.ToSlash(path)
	switch {
	case strings.HasPrefix(p, "/nix/store/"):
		return InstallNix
	case strings.Contains(p, "/node_modules/"):
		return InstallNpm
	case strings.Contains(p, "/Cellar/"),
		strings.HasPrefix(p, "/home/linuxbrew/.linuxbrew/"):
		return InstallBrew
	default:
		return InstallManual
	}
}
