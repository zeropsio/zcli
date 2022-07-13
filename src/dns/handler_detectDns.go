package dns

import (
	"os/exec"
	"regexp"
	"runtime"

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
	LocalDnsManagementWindows        LocalDnsManagement = "WINDOWS"
)

func DetectDns() (LocalDnsManagement, error) {

	binaryLocationExists, err := utils.FileExists(scutil.BinaryLocation)
	if err != nil {
		return "", err
	}
	if binaryLocationExists {
		return LocalDnsManagementScutil, nil
	}

	resolvExists, err := utils.FileExists(constants.ResolvFilePath)
	if err != nil {
		return "", err
	}

	if resolvExists {
		ok, err := isSystemdResolve()
		if err != nil {
			return "", err
		}
		if ok {
			return LocalDnsManagementSystemdResolve, nil
		}
	}

	_, err = exec.LookPath("resolvconf")
	if err == nil {
		return LocalDnsManagementResolveConf, nil
	}

	if resolvExists {
		return LocalDnsManagementFile, nil
	}

	if runtime.GOOS == "windows" {
		return LocalDnsManagementWindows, nil
	}

	return LocalDnsManagementUnknown, nil
}

func isValidSystemdResolveResolveConf(filePath string) (bool, error) {
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

func isSystemdResolve() (bool, error) {

	// resolve.conf is valid for systemd-resolve
	validSystemd, err := isValidSystemdResolveResolveConf(constants.ResolvFilePath)
	if err != nil {
		return false, err
	}
	if !validSystemd {
		return false, nil
	}

	// systemd-resolved unit is running
	if _, err := cmdRunner.Run(exec.Command("pidof", "systemd-resolved")); err != nil {
		return false, nil
	}

	// resolvectl binary exists in PATH
	if _, err := exec.LookPath("resolvectl"); err != nil {
		return false, nil
	}

	return true, nil
}
