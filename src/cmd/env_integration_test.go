//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/constants"
)

// Discovered behavior (out of scope for the kong port per the
// test-first-migration rule): printer.Style(style, text) returns the style's
// empty Value() — not the text — when the process is not attached to a TTY.
// So env-var NAMES wrapped with printer.Style (header, item) disappear in
// CI/pipe mode. Only the i18n descriptions and the lipgloss-rendered TABLE
// remain visible.
//
// The test asserts on what's actually visible: the descriptions and the
// currently-used data file path resolved from the env override.
func TestEnvCommand_PrintsPathsInTable(t *testing.T) {
	f := newFixture(t)

	res := f.RunKong(nil, "env")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "Path to a log file",
		"env-var description should appear in stdout")
	assert.Contains(t, res.Stdout, f.DataPath,
		"current data file path should show in the table")
	_ = constants.CliLogFilePathEnvVar // referenced to keep the import; bug above explains why this var name isn't asserted on
}
