package dns

import (
	"os/exec"
	"regexp"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/scutil"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/cmdRunner"
)

type LocalDnsManagement string

const (
	LocalDnsManagementSystemdResolve LocalDnsManagement = "SYSTEMD_RESOLVE"
	LocalDnsManagementResolveConf    LocalDnsManagement = "RESOLVCONF"
	LocalDnsManagementFile           LocalDnsManagement = "FILE"
	LocalDnsManagementScutil         LocalDnsManagement = "SCUTIL"
	LocalDnsManagementUnknown        LocalDnsManagement = "UNKNOWN"
)

func DetectDns() (LocalDnsManagement, error) {

	if utils.FileExists(scutil.BinaryLocation) {
		return LocalDnsManagementScutil, nil
	}

	resolvExists := utils.FileExists(constants.ResolvFilePath)

	if resolvExists {
		valid, err := isValidSystemdResolve(constants.ResolvFilePath)
		if err != nil {
			return "", err
		}

		if valid {
			_, err := cmdRunner.Run(exec.Command("pidof", "systemd-resolved"))
			if err == nil {
				return LocalDnsManagementSystemdResolve, nil
			}
		}
	}

	_, err := exec.LookPath("resolvconf")
	if err == nil {
		return LocalDnsManagementResolveConf, nil
	}

	if resolvExists {
		return LocalDnsManagementFile, nil
	}

	return LocalDnsManagementUnknown, nil
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
