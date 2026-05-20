package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/constants"
)

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

func TestUpgradeCheckExplicitVersion(t *testing.T) {
	f := newFixture(t)
	// No version-API stub: --version is resolved without contacting the API.
	res := f.Run("upgrade", "--check", "--version", "v1.2.3")

	require.Equalf(t, 1, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "Current: local")
	assert.Contains(t, res.Stdout, "Latest:  v1.2.3")
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
