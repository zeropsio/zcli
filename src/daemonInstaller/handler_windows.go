//go:build windows
// +build windows

package daemonInstaller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/utils"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type windowsRecord struct {
	name         string
	description  string
	dependencies []string
}

func newDaemon(name, description string, dependencies []string) (daemon, error) {
	return &windowsRecord{
		name:         name,
		description:  description,
		dependencies: dependencies,
	}, nil
}

const errnoServiceAlreadyExists = syscall.Errno(1073)

func (daemon *windowsRecord) Install() error {
	exists, err := utils.FileExists(filepath.Join(constants.WireguardPath, "wireguard.exe"))
	if err != nil || !exists {
		return errors.New(i18n.DaemonInstallWireguardNotFound)
	}

	err = checkAndRunAsAdmin()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	ser, err := m.CreateService(daemon.name, binaryPath, mgr.Config{
		Description:      daemon.description,
		DisplayName:      "zerops daemon",
		StartType:        mgr.StartAutomatic,
		ServiceStartName: ".\\LocalSystem",
		ErrorControl:     mgr.ErrorNormal,
	}, "daemon", "run")

	if errors.Is(err, errnoServiceAlreadyExists) {
		ser, err = m.OpenService(daemon.name)
	}

	if err != nil {
		return err
	}
	defer ser.Close()

	return ser.Start()
}

func (daemon *windowsRecord) Remove() error {
	if !daemon.IsInstalled() {
		return ErrNotInstalled
	}

	err := checkAndRunAsAdmin()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	ser, err := m.OpenService(daemon.name)
	if err != nil {
		return err
	}
	defer ser.Close()

	ser.Control(svc.Stop)
	return ser.Delete()
}

func (daemon *windowsRecord) IsInstalled() bool {
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return false
	}
	defer windows.Close(h)

	if err != nil {
		return false
	}
	serviceName, err := syscall.UTF16PtrFromString(daemon.name)
	if err != nil {
		return false
	}
	ser, err := windows.OpenService(h, serviceName, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(ser)
	return true
}

func checkAndRunAsAdmin() error {
	if !runsUnderAdmin() {
		err := runAsAdmin()
		if err != nil {
			return err
		}
		fmt.Println(i18n.DaemonElevated)
		return ErrElevatedPrivileges
	}
	return nil
}

func runAsAdmin() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(append([]string{"/k", exe}, os.Args[1:]...), " ")
	exe = "cmd.exe"

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	return err
}
func runsUnderAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}
