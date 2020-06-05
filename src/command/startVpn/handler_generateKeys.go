package startVpn

import (
	"bytes"
	"os/exec"
)

func (h *Handler) generateKeys() (public, private string, err error) {

	privateKeyOutput, err := h.sudoers.RunCommand(exec.Command("wg", "genkey"))
	if err != nil {
		return
	}
	privateKey := privateKeyOutput[0 : len(privateKeyOutput)-1]

	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = bytes.NewReader(privateKey)
	publicKeyOutput, err := h.sudoers.RunCommand(cmd)
	if err != nil {
		return
	}

	publicKey := publicKeyOutput[0 : len(publicKeyOutput)-1]

	return string(publicKey), string(privateKey), nil
}
