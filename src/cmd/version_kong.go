package cmd

import (
	"context"
	"runtime"

	"github.com/zeropsio/zcli/src/kongCmd"
	getVersion "github.com/zeropsio/zcli/src/version"
)

// VersionCmd prints the build version and a freshness check. Works for both
// guest and authenticated users; no RequireAuth needed.
type VersionCmd struct{}

func (*VersionCmd) Run(ctx context.Context, d *kongCmd.CmdData) error {
	d.Stdout.Printf("zcli version %s (%s) %s/%s\n",
		getVersion.GetCurrent(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
	getVersion.PrintVersionCheck(ctx, d.Stdout)
	return nil
}
