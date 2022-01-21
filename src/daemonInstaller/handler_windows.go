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

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
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

func (daemon *windowsRecord) Install() error {
	exists, err := utils.FileExists(filepath.Join(constants.WireguardPath, "wireguard.exe"))
	if err != nil || !exists {
		return errors.New(i18n.DaemonInstallWireguardNotFound)
	}

	if !amAdmin() {
		err := runMeElevated()
		if err != nil {
			return err
		}
		fmt.Println(i18n.DaemonElevated)
		return nil
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

	ser, err := m.CreateService("zerops", binaryPath, mgr.Config{
		Description:      "zcli zerops daemon",
		DisplayName:      "zerops",
		StartType:        mgr.StartAutomatic,
		ServiceStartName: ".\\LocalSystem",
		ErrorControl:     mgr.ErrorNormal,
	}, "daemon", "run")

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

	if !amAdmin() {
		err := runMeElevated()
		if err != nil {
			return err
		}
		fmt.Println(i18n.DaemonElevated)
		return nil
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	ser, err := m.OpenService("zerops")
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
		fmt.Println(err)
		return false
	}
	defer windows.Close(h)

	if err != nil {
		fmt.Println(err)
		return false
	}
	serviceName, err := syscall.UTF16PtrFromString("zerops")
	if err != nil {
		fmt.Println(err)
		return false
	}
	ser, err := windows.OpenService(h, serviceName, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer windows.CloseServiceHandle(ser)
	return true
}

func runMeElevated() error {
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
func amAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}
