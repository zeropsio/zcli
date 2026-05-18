//go:build devel

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestServiceDeployCommand_HappyPath exercises the deploy variant of the
// push happy path: same scaffolding as push, but the final request goes to
// /app-version/{id}/deploy instead of /build-and-deploy. Shared helpers live
// in pushDeploy_helpers_test.go.
func TestServiceDeployCommand_HappyPath(t *testing.T) {
	f := newFixture(t)
	f.SeedLogin("test-token")
	workDir := t.TempDir()
	writeZeropsYaml(t, workDir, "demo")
	// service deploy archives files relative to working-dir; give it a single
	// file so the tar isn't empty and the upload handler sees bytes.
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "index.html"), []byte("<html/>"), 0o600))

	s := registerDeployStubs(t, f, "demo")

	res := f.Run(
		nil,
		"service", "deploy",
		"--service-id", pushServiceID,
		"--working-dir", workDir,
	)
	assertPushSuccess(t, res, s, "demo")
}
