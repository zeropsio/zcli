//go:build devel

package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	f := newFixture(t)

	res := f.Run(nil, "version")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Truef(t, strings.HasPrefix(res.Stdout, "zcli version "), "unexpected stdout: %q", res.Stdout)
}
