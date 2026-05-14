//go:build devel

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/constants"
)

func TestShowDebugLogs_TailsExistingFile(t *testing.T) {
	f := newFixture(t)

	logPath := os.Getenv(constants.CliLogFilePathEnvVar)
	require.NotEmpty(t, logPath, "harness must set log path env")
	require.NoError(t, os.MkdirAll(filepath.Dir(logPath), 0o700))
	require.NoError(t, os.WriteFile(logPath, []byte("first\nsecond\nthird\n"), 0o600))

	res := f.RunKong(nil, "show-debug-logs")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "second", "expected log line to be tailed")
	assert.Contains(t, res.Stderr, "third", "expected log line to be tailed")
}

// TestShowDebugLogs_EmptyFileExitsZero — narrow assertion because the
// empty-file branch routes its message through UxBlocks.PrintInfo which is
// silenced in non-TTY contexts (tests, pipes, CI), while the populated
// branch writes directly to Stderr. The "Debug logs not found" message is
// only seen interactively today. Logged as discovered behavior; out of
// scope for the kong port per the test-first rule.
func TestShowDebugLogs_EmptyFileExitsZero(t *testing.T) {
	f := newFixture(t)

	logPath := os.Getenv(constants.CliLogFilePathEnvVar)
	require.NoError(t, os.MkdirAll(filepath.Dir(logPath), 0o700))
	require.NoError(t, os.WriteFile(logPath, []byte{}, 0o600))

	res := f.RunKong(nil, "show-debug-logs")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
}
