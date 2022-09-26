//go:build linux
// +build linux

package daemon

import (
	"github.com/zerops-io/zcli/src/constants"
)

func daemonDialAddress() string {
	return "unix:///" + constants.DaemonAddress
}
