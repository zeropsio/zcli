//go:build devel

package cmd

import (
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	f := newFixture(t)

	res := f.Run(nil, "version")

	if res.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if !strings.HasPrefix(res.Stdout, "zcli version ") {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
}
