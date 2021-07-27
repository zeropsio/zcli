// +build windows

package dnsServer

import (
	"context"
	"net"
)

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

// listen and serve
func (h *Handler) Run(_ context.Context) error {
	return nil
}

func (h *Handler) StopForward() {

}

func (h *Handler) SetAddresses(_ net.IP, _ []net.IP, _ net.IP, _ net.IPNet) {
}
