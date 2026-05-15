package version

import "testing"

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
			if got := detectFromPath(tc.path); got != tc.want {
				t.Fatalf("detectFromPath(%q) = %v, want %v", tc.path, got, tc.want)
			}
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
		{"manual", InstallManual, true},
		{"", 0, false},
		{"unknown", 0, false},
		{"AUR", 0, false}, // case-sensitive, no implicit fallback
	}
	for _, tc := range cases {
		got, ok := parseChannel(tc.in)
		if ok != tc.wantOk || got != tc.want {
			t.Errorf("parseChannel(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOk)
		}
	}
}

func TestDetectChannelStamp(t *testing.T) {
	saved := channel
	t.Cleanup(func() { channel = saved })

	channel = "nix"
	if got := Detect(); got != InstallNix {
		t.Errorf("Detect() with channel=nix = %v, want InstallNix", got)
	}
	channel = "brew"
	if got := Detect(); got != InstallBrew {
		t.Errorf("Detect() with channel=brew = %v, want InstallBrew", got)
	}
}

func TestInstallMethodPackageManager(t *testing.T) {
	for _, m := range []InstallMethod{InstallNix, InstallNpm, InstallBrew} {
		if !m.IsPackageManager() {
			t.Errorf("%v should be a package manager", m)
		}
		if m.Hint() == "" {
			t.Errorf("%v should have a hint", m)
		}
	}
	if InstallManual.IsPackageManager() {
		t.Error("InstallManual should not be a package manager")
	}
	if InstallManual.Hint() != "" {
		t.Error("InstallManual should have empty hint")
	}
}
