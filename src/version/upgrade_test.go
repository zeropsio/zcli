package version

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/minio/selfupdate"
)

func TestAssetNameFor(t *testing.T) {
	cases := []struct {
		goos, goarch string
		want         string
	}{
		{"linux", "amd64", "zcli-linux-amd64"},
		{"linux", "386", "zcli-linux-i386"},
		{"darwin", "amd64", "zcli-darwin-amd64"},
		{"darwin", "arm64", "zcli-darwin-arm64"},
		{"windows", "amd64", "zcli-win-x64.exe"},
		{"windows", "386", "zcli-win-x64.exe"}, // windows always maps to the same asset
	}
	for _, tc := range cases {
		if got := assetNameFor(tc.goos, tc.goarch); got != tc.want {
			t.Errorf("assetNameFor(%q, %q) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
}

func TestParseChecksum(t *testing.T) {
	body := `abc123def4567890abcdef1234567890abcdef1234567890abcdef1234567890  zcli-linux-amd64
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef *zcli-darwin-arm64
fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210  checksums.txt
`
	t.Run("standard format", func(t *testing.T) {
		got, err := parseChecksum(body, "zcli-linux-amd64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []byte{0xab, 0xc1, 0x23, 0xde, 0xf4, 0x56, 0x78, 0x90}
		if !bytes.Equal(got[:8], want) {
			t.Errorf("got first 8 bytes %x, want %x", got[:8], want)
		}
	})

	t.Run("starred name", func(t *testing.T) {
		got, err := parseChecksum(body, "zcli-darwin-arm64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 32 {
			t.Errorf("sha256 should be 32 bytes, got %d", len(got))
		}
	})

	t.Run("missing asset", func(t *testing.T) {
		_, err := parseChecksum(body, "zcli-darwin-amd64")
		if err == nil {
			t.Fatal("expected error for missing asset")
		}
	})

	t.Run("malformed lines skipped", func(t *testing.T) {
		bad := "garbage\n\nonefield\nabc123  zcli-linux-amd64\n"
		_, err := parseChecksum(bad, "zcli-linux-amd64")
		if err != nil {
			t.Errorf("expected to find asset past bad lines: %v", err)
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		bad := "zzznot-hex  zcli-linux-amd64\n"
		_, err := parseChecksum(bad, "zcli-linux-amd64")
		if err == nil {
			t.Fatal("expected error for invalid hex")
		}
	})
}

func TestRequireSelfUpdatable(t *testing.T) {
	savedChannel := channel
	t.Cleanup(func() { channel = savedChannel })

	for _, stamp := range []string{"npm", "brew", "nix", "deb"} {
		channel = stamp
		if err := RequireSelfUpdatable(); err == nil {
			t.Errorf("channel %q: expected refusal, got nil", stamp)
		}
	}

	channel = "manual"
	if err := RequireSelfUpdatable(); err != nil {
		t.Errorf("manual channel: expected no refusal, got %v", err)
	}
}

func TestPlanUpgradeAlwaysSucceeds(t *testing.T) {
	savedChannel := channel
	t.Cleanup(func() { channel = savedChannel })

	for _, stamp := range []string{"manual", "npm", "brew", "nix", "deb"} {
		channel = stamp
		plan, err := PlanUpgrade(t.Context(), UpgradeOptions{TargetVersion: "v1.0.0"})
		if err != nil {
			t.Errorf("channel %q: expected plan, got error %v", stamp, err)
			continue
		}
		if plan.Target != "v1.0.0" {
			t.Errorf("channel %q: target = %q, want v1.0.0", stamp, plan.Target)
		}
	}
}

// upgradeFixture wires a fake release server and a recording applyUpdate
// stub. The returned cleanup restores the package-level overrides.
type upgradeFixture struct {
	server   *httptest.Server
	binary   []byte
	checksum [32]byte
	applied  []byte
	applyOpt selfupdate.Options
	applyErr error
}

func newUpgradeFixture(t *testing.T, handler http.HandlerFunc) *upgradeFixture {
	t.Helper()
	fix := &upgradeFixture{
		binary: []byte("pretend this is a zcli binary"),
	}
	fix.checksum = sha256.Sum256(fix.binary)

	if handler == nil {
		handler = fix.defaultHandler
	}
	fix.server = httptest.NewServer(handler)

	savedURL := releasesURL
	savedApply := applyUpdate
	releasesURL = fix.server.URL + "/%s/%s"
	applyUpdate = func(r io.Reader, opts selfupdate.Options) error {
		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		fix.applied = b
		fix.applyOpt = opts
		return fix.applyErr
	}

	t.Cleanup(func() {
		fix.server.Close()
		releasesURL = savedURL
		applyUpdate = savedApply
	})
	return fix
}

func (f *upgradeFixture) defaultHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/"+checksumsName):
		fmt.Fprintf(w, "%x  %s\n", f.checksum, assetName())
	case strings.HasSuffix(r.URL.Path, "/"+assetName()):
		_, _ = w.Write(f.binary)
	default:
		http.NotFound(w, r)
	}
}

func TestUpgradeHappyPath(t *testing.T) {
	fix := newUpgradeFixture(t, nil)

	plan := &UpgradePlan{Current: "v0.9.0", Target: "v1.0.0"}
	if err := Upgrade(context.Background(), plan); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if !bytes.Equal(fix.applied, fix.binary) {
		t.Errorf("applied binary mismatch: got %q, want %q", fix.applied, fix.binary)
	}
	if !bytes.Equal(fix.applyOpt.Checksum, fix.checksum[:]) {
		t.Errorf("checksum passed to selfupdate = %x, want %x", fix.applyOpt.Checksum, fix.checksum[:])
	}
}

func TestUpgradeAssetNotListed(t *testing.T) {
	fix := newUpgradeFixture(t, nil)
	// Override the handler to omit the platform asset from checksums.txt.
	fix.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/"+checksumsName):
			fmt.Fprintln(w, "deadbeef  some-other-asset")
		default:
			http.NotFound(w, r)
		}
	})

	plan := &UpgradePlan{Current: "v0.9.0", Target: "v1.0.0"}
	err := Upgrade(context.Background(), plan)
	if err == nil {
		t.Fatal("expected error when asset is missing from checksums.txt")
	}
	if !strings.Contains(err.Error(), "not listed") {
		t.Errorf("error message %q should mention the missing asset", err)
	}
}

func TestUpgradeChecksumsUnreachable(t *testing.T) {
	newUpgradeFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))

	plan := &UpgradePlan{Current: "v0.9.0", Target: "v1.0.0"}
	err := Upgrade(context.Background(), plan)
	if err == nil {
		t.Fatal("expected error when checksums.txt fetch fails")
	}
	if !strings.Contains(err.Error(), "checksums.txt") {
		t.Errorf("error %q should mention checksums.txt", err)
	}
}

func TestUpgradeBinaryNotFound(t *testing.T) {
	fix := newUpgradeFixture(t, nil)
	// Serve checksums.txt but 404 the binary.
	fix.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/"+checksumsName) {
			fmt.Fprintf(w, "%x  %s\n", fix.checksum, assetName())
			return
		}
		http.NotFound(w, r)
	})

	plan := &UpgradePlan{Current: "v0.9.0", Target: "v1.0.0"}
	err := Upgrade(context.Background(), plan)
	if err == nil {
		t.Fatal("expected error when binary fetch fails")
	}
	if !strings.Contains(err.Error(), "download binary") {
		t.Errorf("error %q should mention binary download", err)
	}
}

func TestUpgradeApplyError(t *testing.T) {
	fix := newUpgradeFixture(t, nil)
	fix.applyErr = fmt.Errorf("permission denied: cannot replace binary")

	plan := &UpgradePlan{Current: "v0.9.0", Target: "v1.0.0"}
	err := Upgrade(context.Background(), plan)
	if err == nil {
		t.Fatal("expected error when apply fails")
	}
	if !strings.Contains(err.Error(), "sudo zcli upgrade") {
		t.Errorf("permission errors should suggest sudo, got %q", err)
	}
}
