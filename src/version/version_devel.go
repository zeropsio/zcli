//go:build devel

package version

import (
	"context"
	_ "embed"

	"github.com/zeropsio/zcli/src/printer"
)

var version = "local"

func IsVersionCheckMismatch(context.Context) bool {
	return false
}

func GetVersionCheckMismatch() (string, error) {
	return "", nil
}

func PrintVersionCheck(context.Context, printer.Printer) {}

func GetCurrent() string {
	return version
}
