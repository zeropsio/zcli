package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/constants"
)

// stubVersion overrides what upgrade.NewUpgrader() reports as the current
// version. The test binary is built by `go test` (no ldflag stamp), and
// NewUpgrader honors VersionEnvVar before falling back to the stamped value,
// so t.Setenv is the whole hook - no test-only helper in the upgrade pkg.
func (f *fixture) stubVersion(v string) {
	f.t.Setenv(constants.VersionEnvVar, v)
}

// stubVersionAPI points ZEROPS_VERSION_API_URL at a handler on the fixture
// server that returns the given status and, when tagName is non-empty, a
// GitHub-release-style body. The current binary reports version "local" in
// tests (the version var is only stamped at release build time).
func (f *fixture) stubVersionAPI(status int, tagName string) {
	const path = "/release/latest"
	f.Mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		if tagName != "" {
			_, _ = fmt.Fprintf(w, `{"tag_name":%q}`, tagName)
		}
	})
	f.t.Setenv(constants.VersionApiUrlEnvVar, f.Server.URL+path)
}

func TestUpgradeCheckUpToDate(t *testing.T) {
	f := newFixture(t)
	f.stubVersionAPI(http.StatusOK, "local") // latest == current

	res := f.Run("upgrade", "--check")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "Current: local")
	assert.Contains(t, res.Stdout, "Latest:  local")
}

func TestUpgradeCheckBehind(t *testing.T) {
	f := newFixture(t)
	f.stubVersionAPI(http.StatusOK, "v2.0.0")

	res := f.Run("upgrade", "--check")

	require.Equalf(t, 1, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "Latest:  v2.0.0")
}

// A local PROD build stamps git-describe info as semver build metadata
// (Makefile PROD_VERSION rewrites `-N-gHASH` to `+N.gHASH`). semver.Compare
// ignores the `+...` suffix, so being 11 commits ahead of v1.1.0 should
// report as up to date, not as "behind v1.1.0".
func TestUpgradeCheckAheadOfTagViaBuildMetadata(t *testing.T) {
	f := newFixture(t)
	f.stubVersion("v1.1.0+11.g03aedf4")
	f.stubVersionAPI(http.StatusOK, "v1.1.0")

	res := f.Run("upgrade", "--check")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q stdout=%q", res.Stderr, res.Stdout)
	assert.Contains(t, res.Stdout, "Current: v1.1.0+11.g03aedf4")
	assert.Contains(t, res.Stdout, "Latest:  v1.1.0")
}

func TestUpgradeCheckExplicitVersion(t *testing.T) {
	f := newFixture(t)
	f.stubReleaseTag("v1.2.3")
	// No version-API stub: --version is resolved without contacting the API.
	res := f.Run("upgrade", "--check", "--version", "v1.2.3")

	require.Equalf(t, 1, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "Current: local")
	assert.Contains(t, res.Stdout, "Latest:  v1.2.3")
}

// PlanUpgrade HEADs the binary URL when --version is set; a 404 surfaces as
// "release does not exist" with exit 2 in --check mode.
func TestUpgradeCheckInvalidTag(t *testing.T) {
	f := newFixture(t)
	// No stubReleaseTag for vBOGUS - fixture's default 404 stands in for
	// GitHub returning 404 on a typo'd tag.

	res := f.Run("upgrade", "--check", "--version", "vBOGUS")

	require.Equalf(t, 2, res.ExitCode, "stdout=%q stderr=%q", res.Stdout, res.Stderr)
	assert.Contains(t, res.Stderr, "vBOGUS")
	assert.Contains(t, res.Stderr, "does not exist")
}

// --check --version vX on a pre-v1.1.0 tag exits with code 3 (target known
// but unreachable by `zcli upgrade`) and prints the install.sh hint to
// stderr, so scripts can distinguish "use install.sh" from "actual upgrade
// available" (exit 1).
func TestUpgradeCheckTargetUnreachable(t *testing.T) {
	f := newFixture(t)
	f.stubReleaseTag("v1.0.67")

	res := f.Run("upgrade", "--check", "--version", "v1.0.67")

	require.Equalf(t, 3, res.ExitCode, "stdout=%q stderr=%q", res.Stdout, res.Stderr)
	assert.Contains(t, res.Stdout, "Latest:  v1.0.67")
	assert.Contains(t, res.Stderr, "install.sh")
	assert.Contains(t, res.Stderr, "v1.0.67")
}

// Releases before v1.1.0 didn't ship a checksums.txt, so the self-upgrader
// would 404 mid-way. Apply is gated by Plan.RequireSelfUpgradable; the
// command must point users at install.sh for those older tags.
func TestUpgradeRejectsPreRework(t *testing.T) {
	f := newFixture(t)
	f.stubReleaseTag("v1.0.67")

	res := f.Run("upgrade", "--yes", "--version", "v1.0.67")

	require.NotEqualf(t, 0, res.ExitCode, "stdout=%q stderr=%q", res.Stdout, res.Stderr)
	combined := res.Stdout + res.Stderr
	assert.Contains(t, combined, "v1.0.67")
	assert.Contains(t, combined, "install.sh")
}

func TestUpgradeInvalidDownloadTimeout(t *testing.T) {
	f := newFixture(t)
	f.stubVersionAPI(http.StatusOK, "v2.0.0")

	res := f.Run("upgrade", "--yes", "--download-timeout", "not-a-duration")

	require.NotEqualf(t, 0, res.ExitCode, "stdout=%q stderr=%q", res.Stdout, res.Stderr)
	assert.Contains(t, res.Stderr, "--download-timeout")
}

func TestUpgradeCheckError(t *testing.T) {
	f := newFixture(t)
	f.stubVersionAPI(http.StatusInternalServerError, "")

	res := f.Run("upgrade", "--check")

	require.Equalf(t, 2, res.ExitCode, "stdout=%q stderr=%q", res.Stdout, res.Stderr)
	assert.Contains(t, res.Stderr, "error")
}

func TestUpgradeAlreadyOnLatest(t *testing.T) {
	f := newFixture(t)
	f.stubVersionAPI(http.StatusOK, "local") // latest == current, no --version given

	res := f.Run("upgrade", "--yes")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "already on local")
}
