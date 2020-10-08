package vpn

import (
	"github.com/zerops-io/zcli/src/daemonStorage"
)

func (h *Handler) StopVpn() (err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	defer func() {
		if err != nil {
			h.logger.Error(err)
		}
	}()

	err = h.cleanVpn()
	if err != nil {
		return err
	}

	localDnsManagement, err := h.detectDns()
	if err != nil {
		return err
	}

	data := h.storage.Data()
	err = h.cleanDns(data.DnsIp, data.ClientIp, localDnsManagement)
	if err != nil {
		return err
	}

	dataReset := &daemonStorage.Data{}
	err = h.storage.Save(dataReset)
	if err != nil {
		return err
	}

	return nil
}
