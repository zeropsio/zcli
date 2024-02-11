package cmdRunner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	err := &execError{
		prev:     ErrOperationNotPermitted,
		exitCode: 1,
	}

	require.ErrorIs(t, err, ErrOperationNotPermitted)
	require.NotErrorIs(t, err, ErrCannotFindDevice)
	require.Equal(t, 1, err.ExitCode())
}
