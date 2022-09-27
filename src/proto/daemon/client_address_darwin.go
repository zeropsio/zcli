//go:build darwin
// +build darwin

package daemon

import (
	"github.com/zeropsio/zcli/src/constants"
)

func daemonDialAddress() string {
	return "unix:///" + constants.DaemonAddress
}
