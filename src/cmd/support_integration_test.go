//go:build devel

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportCommand_PrintsContactInfo(t *testing.T) {
	f := newFixture(t)

	res := f.Run(nil, "support")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout, "team@zerops.io", "email contact must appear")
	assert.Contains(t, res.Stdout, "discord.com/invite", "discord link must appear")
}
