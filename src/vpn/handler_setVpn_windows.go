//go:build windows
// +build windows

package vpn

import (
	"errors"

	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

func (h *Handler) setVpn(selectedVpnAddress, privateKey string, mtu uint32, response *zeropsVpnProtocol.StartVpnResponse) error {
	return errors.New("windows is not supported")
}
