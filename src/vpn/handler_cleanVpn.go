package vpn

import (
	"github.com/zerops-io/zcli/src/wgquick"
)

func (h *Handler) cleanVpn() error {

	var err error

	h.logger.Debug("clean vpn start")

	err = wgquick.New().Down("zerops")
	if err != nil {
		return err
	}

	h.logger.Debug("clean vpn end")

	return nil
}
