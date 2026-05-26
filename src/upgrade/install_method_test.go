package upgrade

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectFromPath(t *testing.T) {
	cases := []struct {
		name string
		path string
		want InstallMethod
	}{
		{"nix store", "/nix/store/abc123-zcli/bin/zcli", InstallNix},
		{"npm global on macos", "/Users/x/.npm-global/lib/node_modules/@zerops/zcli/utils/bin/zcli", InstallNpm},
		{"npm via nvm", "/Users/x/.nvm/versions/node/v22.0.0/lib/node_modules/@zerops/zcli/utils/bin/zcli", InstallNpm},
		{"homebrew apple silicon", "/opt/homebrew/Cellar/zcli/1.2.3/bin/zcli", InstallBrew},
		{"homebrew intel", "/usr/local/Cellar/zcli/1.2.3/bin/zcli", InstallBrew},
		{"linuxbrew", "/home/linuxbrew/.linuxbrew/Cellar/zcli/1.2.3/bin/zcli", InstallBrew},
		{"install.sh path", "/usr/local/bin/zcli", InstallManual},
		{"home bin", "/Users/x/.local/bin/zcli", InstallManual},
		{"windows install path", `C:\Program Files\zcli\zcli.exe`, InstallManual},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equalf(t, tc.want, detectFromPath(tc.path), "detectFromPath(%q)", tc.path)
		})
	}
}

func TestParseChannel(t *testing.T) {
	cases := []struct {
		in     string
		want   InstallMethod
		wantOk bool
	}{
		{"nix", InstallNix, true},
		{"npm", InstallNpm, true},
		{"brew", InstallBrew, true},
		{"deb", InstallDeb, true},
		{"manual", InstallManual, true},
		{"", 0, false},
		{"unknown", 0, false},
		{"AUR", 0, false}, // case-sensitive, no implicit fallback
	}
	for _, tc := range cases {
		got, ok := parseChannel(tc.in)
		assert.Equalf(t, tc.wantOk, ok, "parseChannel(%q) ok", tc.in)
		assert.Equalf(t, tc.want, got, "parseChannel(%q) method", tc.in)
	}
}

func TestDetectChannelStamp(t *testing.T) {
	assert.Equal(t, InstallNix, (Upgrader{channel: "nix"}).Detect(), "channel=nix")
	assert.Equal(t, InstallBrew, (Upgrader{channel: "brew"}).Detect(), "channel=brew")
}

func TestInstallMethodPackageManager(t *testing.T) {
	for _, m := range []InstallMethod{InstallNix, InstallNpm, InstallBrew, InstallDeb, InstallManual} {
		assert.NotEmptyf(t, m.Hint(), "%v should have a non-empty hint", m)
	}
	for _, m := range []InstallMethod{InstallNix, InstallNpm, InstallBrew, InstallDeb} {
		assert.Truef(t, m.IsPackageManager(), "%v should be a package manager", m)
	}
	assert.False(t, InstallManual.IsPackageManager(), "InstallManual should not be a package manager")
}
