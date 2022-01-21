package vpn

import (
	"bytes"
	"os/exec"

	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

func (h *Handler) generateKeys() (public, private string, err error) {

	h.logger.Debug("generate keys start")

	privateKeyOutput, err := cmdRunner.Run(exec.Command("wg", "genkey"))
	if err != nil {
		return
	}
	privateKey := bytes.TrimSpace(privateKeyOutput)

	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = bytes.NewReader(privateKey)
	publicKeyOutput, err := cmdRunner.Run(cmd)
	if err != nil {
		return
	}

	h.logger.Debug("generate keys end")

	publicKey := bytes.TrimSpace(publicKeyOutput)

	return string(publicKey), string(privateKey), nil
}
