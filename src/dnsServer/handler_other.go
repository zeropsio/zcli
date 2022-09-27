//go:build windows || linux
// +build windows linux

package dnsServer

import (
	"context"
	"net"

	"github.com/zeropsio/zcli/src/utils/logger"
)

type Handler struct {
}

func New(_ logger.Logger) *Handler {
	return &Handler{}
}

func (h *Handler) Run(_ context.Context) error {
	return nil
}

func (h *Handler) StopForward() {}

func (h *Handler) SetAddresses(_ net.IP, _ []net.IP, _ net.IP, _ net.IPNet) {
}
