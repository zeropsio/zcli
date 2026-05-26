package upgrade

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/minio/selfupdate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equalf(t, tc.want, assetNameFor(tc.goos, tc.goarch), "assetNameFor(%q, %q)", tc.goos, tc.goarch)
	}
}

func TestParseChecksum(t *testing.T) {
	body := `abc123def4567890abcdef1234567890abcdef1234567890abcdef1234567890  zcli-linux-amd64
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef *zcli-darwin-arm64
fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210  checksums.txt
`
	t.Run("standard format", func(t *testing.T) {
		got, err := parseChecksum(body, "zcli-linux-amd64")
		require.NoError(t, err)
		want := []byte{0xab, 0xc1, 0x23, 0xde, 0xf4, 0x56, 0x78, 0x90}
		assert.Equal(t, want, got[:8], "first 8 bytes")
	})

	t.Run("starred name", func(t *testing.T) {
		got, err := parseChecksum(body, "zcli-darwin-arm64")
		require.NoError(t, err)
		assert.Len(t, got, 32, "sha256 should be 32 bytes")
	})

	t.Run("missing asset", func(t *testing.T) {
		_, err := parseChecksum(body, "zcli-darwin-amd64")
		require.Error(t, err, "expected error for missing asset")
	})

	t.Run("malformed lines skipped", func(t *testing.T) {
		bad := "garbage\n\nonefield\nabc123  zcli-linux-amd64\n"
		_, err := parseChecksum(bad, "zcli-linux-amd64")
		assert.NoError(t, err, "expected to find asset past bad lines")
	})

	t.Run("invalid hex", func(t *testing.T) {
		bad := "zzznot-hex  zcli-linux-amd64\n"
		_, err := parseChecksum(bad, "zcli-linux-amd64")
		require.Error(t, err, "expected error for invalid hex")
	})
}

func TestRequireSelfUpdatable(t *testing.T) {
	for _, stamp := range []string{"npm", "brew", "nix", "deb"} {
		assert.Errorf(t, (Upgrader{channel: stamp}).RequireSelfUpdatable(), "channel %q: expected refusal", stamp)
	}
	assert.NoError(t, (Upgrader{channel: "manual"}).RequireSelfUpdatable(), "manual channel: expected no refusal")
}

func TestPlanUpgradeAlwaysSucceeds(t *testing.T) {
	for _, stamp := range []string{"manual", "npm", "brew", "nix", "deb"} {
		plan, err := (Upgrader{channel: stamp}).PlanUpgrade(t.Context(), Options{TargetVersion: "v1.0.0"})
		if !assert.NoErrorf(t, err, "channel %q: expected plan", stamp) {
			continue
		}
		assert.Equalf(t, "v1.0.0", plan.target, "channel %q: target", stamp)
	}
}

// upgradeFixture wires a fake release server and an Upgrader whose apply
// func captures the bytes instead of replacing the running test binary.
type upgradeFixture struct {
	server   *httptest.Server
	upgrader Upgrader
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

	fix.upgrader = Upgrader{
		releasesURL:     fix.server.URL + "/%s/%s",
		downloadTimeout: defaultDownloadTimeout,
		apply: func(r io.Reader, opts selfupdate.Options) error {
			b, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			fix.applied = b
			fix.applyOpt = opts
			return fix.applyErr
		},
	}

	t.Cleanup(fix.server.Close)
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

	plan := Plan{current: "v0.9.0", target: "v1.0.0"}
	require.NoError(t, fix.upgrader.Apply(context.Background(), plan))
	assert.Equal(t, fix.binary, fix.applied, "applied binary")
	assert.Equal(t, fix.checksum[:], fix.applyOpt.Checksum, "checksum passed to selfupdate")
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

	plan := Plan{current: "v0.9.0", target: "v1.0.0"}
	err := fix.upgrader.Apply(context.Background(), plan)
	require.Error(t, err, "expected error when asset is missing from checksums.txt")
	assert.Contains(t, err.Error(), "not listed", "error should mention the missing asset")
}

func TestUpgradeChecksumsUnreachable(t *testing.T) {
	fix := newUpgradeFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))

	plan := Plan{current: "v0.9.0", target: "v1.0.0"}
	err := fix.upgrader.Apply(context.Background(), plan)
	require.Error(t, err, "expected error when checksums.txt fetch fails")
	assert.Contains(t, err.Error(), "checksums.txt")
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

	plan := Plan{current: "v0.9.0", target: "v1.0.0"}
	err := fix.upgrader.Apply(context.Background(), plan)
	require.Error(t, err, "expected error when binary fetch fails")
	assert.Contains(t, err.Error(), "download binary")
}

func TestUpgradeApplyError(t *testing.T) {
	fix := newUpgradeFixture(t, nil)
	fix.applyErr = fmt.Errorf("permission denied: cannot replace binary")

	plan := Plan{current: "v0.9.0", target: "v1.0.0"}
	err := fix.upgrader.Apply(context.Background(), plan)
	require.Error(t, err, "expected error when apply fails")
	assert.Contains(t, err.Error(), "sudo zcli upgrade", "permission errors should suggest sudo")
}

func TestAvailableReleases(t *testing.T) {
	// One payload exercises every branch:
	//   * v3.0.0 draft - always dropped
	//   * latest-stable non-semver - always dropped
	//   * v2.0.0 stable - kept
	//   * v1.5.0-rc.1 prerelease via semver suffix - kept only with includePrerelease
	//   * v1.4.0 with GitHub prerelease flag - kept only with includePrerelease
	//   * v1.1.0 stable post-rework
	//   * v1.0.67 stable pre-rework
	payload := `[
		{"tag_name":"v3.0.0","draft":true},
		{"tag_name":"latest-stable","draft":false},
		{"tag_name":"v2.0.0","draft":false},
		{"tag_name":"v1.5.0-rc.1","draft":false},
		{"tag_name":"v1.4.0","draft":false,"prerelease":true},
		{"tag_name":"v1.1.0","draft":false},
		{"tag_name":"v1.0.67","draft":false}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, payload)
	}))
	t.Cleanup(srv.Close)

	u := Upgrader{releasesListURL: srv.URL}

	t.Run("default drops prereleases", func(t *testing.T) {
		releases, err := u.AvailableReleases(context.Background(), false)
		require.NoError(t, err)
		assert.Equal(t, []Release{
			{Tag: "v2.0.0", SelfUpgradable: true, Prerelease: false},
			{Tag: "v1.1.0", SelfUpgradable: true, Prerelease: false},
			{Tag: "v1.0.67", SelfUpgradable: false, Prerelease: false},
		}, releases)
	})

	t.Run("includePrerelease keeps -rc and flagged tags", func(t *testing.T) {
		releases, err := u.AvailableReleases(context.Background(), true)
		require.NoError(t, err)
		assert.Equal(t, []Release{
			{Tag: "v2.0.0", SelfUpgradable: true, Prerelease: false},
			{Tag: "v1.5.0-rc.1", SelfUpgradable: true, Prerelease: true},
			{Tag: "v1.4.0", SelfUpgradable: true, Prerelease: true},
			{Tag: "v1.1.0", SelfUpgradable: true, Prerelease: false},
			{Tag: "v1.0.67", SelfUpgradable: false, Prerelease: false},
		}, releases)
	})
}
