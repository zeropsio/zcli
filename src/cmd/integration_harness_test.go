// Shared scaffolding for the package's integration tests.
//
// Integration tests drive the CLI in-process through cmdBuilder.RunRootCmd,
// point the REST client at an httptest.Server, and isolate per-test state by
// setting ZEROPS_CLI_DATA_FILE_PATH, ZEROPS_CLI_LOG_FILE_PATH, and
// ZEROPS_CLI_YAML_FILE_PATH to per-test temp files. ZEROPS_VERSION_API_URL is
// pointed at the test server so the background version check never reaches the
// real API. yamlReader's package-level cache is reset before and after each
// test. See pushDeploy_helpers_test.go for push/deploy-specific helpers.

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/profile"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/secrets"
	"github.com/zeropsio/zcli/src/yamlReader"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// fixture wires an httptest.Server, an isolated profile + secrets path, a fake repo dir, and
// a preconfigured RegionItem pointing at the test server. Use Run to drive the CLI in-process
// and inspect stdout/stderr/exit code.
type fixture struct {
	t           *testing.T
	Server      *httptest.Server
	Mux         *http.ServeMux
	DataPath    string
	SecretsPath string
	RepoDir     string
	Region      region.Item
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// The yamlReader package caches zerops.yaml bytes in a package-level
	// variable; reset before and after each test to keep fixtures isolated.
	yamlReader.ResetCache()
	t.Cleanup(yamlReader.ResetCache)

	dir := t.TempDir()
	dataPath := filepath.Join(dir, "cli.data")
	secretsPath := filepath.Join(dir, "cli.secrets")
	logPath := filepath.Join(dir, "zcli.log")
	yamlPath := filepath.Join(dir, "zcli.yml")
	// Stand up a fake repo (just a .git directory is enough for DiscoverRepoRoot) so tests
	// that want a repo profile have somewhere to anchor it. Tests that only need global state
	// just ignore RepoDir.
	repoDir := filepath.Join(dir, "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755))
	t.Setenv(constants.CliDataFilePathEnvVar, dataPath)
	t.Setenv(constants.CliLogFilePathEnvVar, logPath)
	t.Setenv(constants.CliZcliYamlFilePathEnvVar, yamlPath)
	t.Setenv(constants.CliTokenEnvVar, "")
	t.Setenv(constants.CliCwdEnvVar, repoDir)
	// Force the plaintext secrets backend with a per-test file so tests never touch the host
	// keyring (which would either fail in CI or, worse, pollute the developer's real keyring).
	t.Setenv(constants.SecretsBackendEnvVar, string(secrets.ModePlaintext))
	t.Setenv(constants.CliSecretsFilePathEnvVar, secretsPath)
	// Keep the background version check off the real network. Tests that
	// exercise the version API re-point this at a registered handler.
	t.Setenv(constants.VersionApiUrlEnvVar, server.URL+"/__version_check__")
	// Route release-asset lookups (used by tag verification on --version)
	// at the test server. Tests opt in to specific tags via stubReleaseTag;
	// unregistered tags hit the default 404 handler.
	t.Setenv(constants.ReleasesURLEnvVar, server.URL+"/__releases__/%s/%s")

	return &fixture{
		t:           t,
		Server:      server,
		Mux:         mux,
		DataPath:    dataPath,
		SecretsPath: secretsPath,
		RepoDir:     repoDir,
		Region: region.Item{
			Name:      "prg1",
			IsDefault: true,
			Address:   server.URL,
		},
	}
}

// SeedLogin writes a profile file with the fixture's region and stores the token in the
// plaintext secrets file at the matching key, simulating a previously completed `zcli login`.
func (f *fixture) SeedLogin(token string) {
	f.t.Helper()
	f.writeProfile(profile.Data{Region: f.Region})
	f.writeSecrets(map[string]string{
		secrets.Key(secrets.GlobalScope, "token"): token,
	})
}

// SeedScopedLogin is SeedLogin plus a persisted project scope in the fixture's repo profile,
// simulating a previously completed `zcli scope project ...` inside that repo.
func (f *fixture) SeedScopedLogin(token, projectID string) {
	f.t.Helper()
	f.writeProfile(profile.Data{Region: f.Region})
	f.writeSecrets(map[string]string{
		secrets.Key(secrets.GlobalScope, "token"): token,
	})
	f.writeRepoProfile(profile.RepoProfile{
		Name:           profile.DefaultRepoProfileName,
		Id:             "test-repo-id",
		ScopeProjectId: uuid.NewProjectIdNullFromString(projectID),
	})
}

func (f *fixture) writeProfile(d profile.Data) {
	// Stamp the current version so Manager.load treats this as already-current and skips the
	// legacy v0 decode path (which would interpret a missing "version" key as "this is the old
	// PascalCase shape" and silently produce empty fields).
	d.StampVersion(profile.CurrentDataVersion)
	b, err := json.Marshal(d)
	require.NoError(f.t, err, "marshal seed profile")
	require.NoError(f.t, os.WriteFile(f.DataPath, b, 0o600), "write seed profile")
}

func (f *fixture) writeSecrets(m map[string]string) {
	b, err := json.Marshal(m)
	require.NoError(f.t, err, "marshal seed secrets")
	require.NoError(f.t, os.WriteFile(f.SecretsPath, b, 0o600), "write seed secrets")
}

// writeRepoProfile seeds a v2 repo file containing exactly one profile (named "default") that
// matches the passed entry. Tests that exercise multi-profile state should write the file
// directly rather than going through this helper.
func (f *fixture) writeRepoProfile(d profile.RepoProfile) {
	if d.Name == "" {
		d.Name = profile.DefaultRepoProfileName
	}
	v := profile.CurrentRepoVersion
	file := map[string]any{
		"version":        v,
		"active_profile": d.Name,
		"profiles":       []profile.RepoProfile{d},
	}
	path := profile.RepoProfilePath(f.RepoDir)
	require.NoError(f.t, os.MkdirAll(filepath.Dir(path), 0o755))
	b, err := json.Marshal(file)
	require.NoError(f.t, err, "marshal seed repo profile")
	require.NoError(f.t, os.WriteFile(path, b, 0o600), "write seed repo profile")
}

// stubReleaseTag tells the fixture's release-asset endpoint to return 200
// for any path under the given tag, so upgrade.PlanUpgrade's tag-existence
// HEAD check succeeds. Unregistered tags fall through to the default mux
// 404, which is what `release does not exist` lookups expect.
func (f *fixture) stubReleaseTag(tag string) {
	f.Mux.HandleFunc(fmt.Sprintf("/__releases__/%s/", tag), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// stubLatestCache seeds the on-disk version cache so
// upgrade.Upgrader.CachedLatest() returns the given tag without contacting
// the API. The cache file layout (cacheEntry / apiResponse) lives in the
// upgrade package and isn't exported; reproducing the JSON shape here keeps
// integration tests independent of that internal type. If that shape
// changes, this helper needs to follow.
func (f *fixture) stubLatestCache(tag string) {
	f.t.Helper()
	cachePath := filepath.Join(filepath.Dir(f.DataPath), constants.VersionCacheFileName)
	entry := map[string]any{
		"fetched_at": time.Now().Format(time.RFC3339Nano),
		"response":   map[string]any{"tag_name": tag},
	}
	b, err := json.Marshal(entry)
	require.NoError(f.t, err, "marshal version cache")
	require.NoError(f.t, os.WriteFile(cachePath, b, 0o644), "write version cache")
}

// HandleJSON registers an exact-path handler returning the given status and
// JSON-encoded body.
func (f *fixture) HandleJSON(path string, status int, body any) {
	f.Mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	})
}

type result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes the CLI in-process with the given args, using the test's
// context. Use RunCtx when a specific (e.g. canceled) context is needed.
func (f *fixture) Run(args ...string) result {
	f.t.Helper()
	return f.RunCtx(f.t.Context(), args...)
}

// RunCtx is Run with an explicit context.
func (f *fixture) RunCtx(ctx context.Context, args ...string) result {
	f.t.Helper()
	var stdout, stderr bytes.Buffer
	code := cmdBuilder.RunRootCmd(
		ctx,
		rootCmd(),
		cmdBuilder.WithArgs(args),
		cmdBuilder.WithStdout(&stdout),
		cmdBuilder.WithStderr(&stderr),
	)
	return result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
}

// LoadStorage reads back the persisted profile JSON — useful for asserting that a command
// updated non-sensitive state on disk. Sensitive values are NOT here; use LoadSecrets for those.
func (f *fixture) LoadStorage() profile.Data {
	f.t.Helper()
	b, err := os.ReadFile(f.DataPath)
	require.NoError(f.t, err, "read storage")
	var d profile.Data
	require.NoError(f.t, json.Unmarshal(b, &d), "unmarshal storage")
	return d
}

// LoadSecrets reads back the persisted secrets file (plaintext mode, set by newFixture).
// Returns the raw key→value map so tests can assert on specific entries.
func (f *fixture) LoadSecrets() map[string]string {
	f.t.Helper()
	b, err := os.ReadFile(f.SecretsPath)
	if os.IsNotExist(err) {
		return map[string]string{}
	}
	require.NoError(f.t, err, "read secrets")
	out := map[string]string{}
	require.NoError(f.t, json.Unmarshal(b, &out), "unmarshal secrets")
	return out
}
