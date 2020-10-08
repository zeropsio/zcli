package vpn

import (
	"os/exec"
	"regexp"

	"github.com/zerops-io/zcli/src/scutil"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

func (h *Handler) detectDns() (localDnsManagement, error) {

	if utils.FileExists(scutil.BinaryLocation) {
		return scutilDnsManagementFile, nil
	}

	if utils.FileExists(resolvFilePath) {
		valid, err := isValidSystemdResolve(resolvFilePath)
		if err != nil {
			return "", err
		}

		if valid {
			_, err := cmdRunner.Run(exec.Command("pidof", "systemd-resolved"))
			if err == nil {
				return localDnsManagementSystemdResolve, nil
			}
		}
	}

	_, err := exec.LookPath("resolvconf")
	if err == nil {
		return localDnsManagementResolveConf, nil
	}

	return localDnsManagementFile, nil
}

func isValidSystemdResolve(filePath string) (bool, error) {
	lines, err := utils.ReadLines(filePath)
	if err != nil {
		return false, err
	}

	nameserverLine := regexp.MustCompile(`[ ]*nameserver[ ]+(.+)`)

	for _, line := range lines {
		submatches := nameserverLine.FindStringSubmatch(line)
		if len(submatches) == 2 {
			if submatches[1] != "127.0.0.53" {
				return false, nil
			} else {
				return true, nil
			}
		}
	}

	return false, nil
}
