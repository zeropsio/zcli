//go:build windows
// +build windows

package daemon

import (
	"github.com/zeropsio/zcli/src/constants"
)

func daemonDialAddress() string {
	return "localhost" + constants.DaemonAddress
}
