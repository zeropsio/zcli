//go:build devel

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/region"
	"github.com/zeropsio/zcli/src/yamlReader"
)

// fixture wires an httptest.Server, an isolated cliStorage path, and a
// preconfigured RegionItem pointing at the test server. Use Run to drive the
// CLI in-process and inspect stdout/stderr/exit code.
type fixture struct {
	t        *testing.T
	Server   *httptest.Server
	Mux      *http.ServeMux
	DataPath string
	Region   region.Item
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
	logPath := filepath.Join(dir, "zcli.log")
	yamlPath := filepath.Join(dir, "zcli.yml")
	t.Setenv(constants.CliDataFilePathEnvVar, dataPath)
	t.Setenv(constants.CliLogFilePathEnvVar, logPath)
	t.Setenv(constants.CliZcliYamlFilePathEnvVar, yamlPath)
	t.Setenv(constants.CliTokenEnvVar, "")

	return &fixture{
		t:        t,
		Server:   server,
		Mux:      mux,
		DataPath: dataPath,
		Region: region.Item{
			Name:      "prg1",
			IsDefault: true,
			Address:   server.URL,
		},
	}
}

// SeedLogin writes a cliStorage file with the given token and the fixture's
// region item, simulating a previously completed `zcli login`.
func (f *fixture) SeedLogin(token string) {
	f.t.Helper()
	data := cliStorage.Data{Token: token, RegionData: f.Region}
	b, err := json.Marshal(data)
	require.NoError(f.t, err, "marshal seed data")
	require.NoError(f.t, os.WriteFile(f.DataPath, b, 0o600), "write seed data")
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

// Run executes the CLI in-process with the given args. ctx defaults to
// context.Background() when nil.
func (f *fixture) Run(ctx context.Context, args ...string) result {
	f.t.Helper()
	if ctx == nil {
		ctx = context.Background()
	}
	var stdout, stderr bytes.Buffer
	code := cmdBuilder.RunRootCmd(rootCmd(), cmdBuilder.RunOptions{
		Ctx:    ctx,
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	return result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
}

// LoadStorage reads back the persisted cliStorage data — useful for asserting
// that a command updated state on disk.
func (f *fixture) LoadStorage() cliStorage.Data {
	f.t.Helper()
	b, err := os.ReadFile(f.DataPath)
	require.NoError(f.t, err, "read storage")
	var d cliStorage.Data
	require.NoError(f.t, json.Unmarshal(b, &d), "unmarshal storage")
	return d
}
