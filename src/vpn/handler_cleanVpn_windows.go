//go:build windows
// +build windows

package vpn

import (
	"errors"
)

func (h *Handler) cleanVpn() error {
	return errors.New("windows is not supported")
}
