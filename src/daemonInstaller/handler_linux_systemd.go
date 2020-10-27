// +build linux

package daemonInstaller

import (
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/zerops-io/zcli/src/constants"
)

const (
	installDir = "/usr/sbin/"
)

type systemDRecord struct {
	name         string
	description  string
	dependencies []string
}

func (daemon *systemDRecord) Install() error {
	if daemon.IsInstalled() {
		return ErrAlreadyInstalled
	}

	logDir, _ := path.Split(constants.LogFilePath)
	daemonStorageDir, _ := path.Split(constants.DaemonStorageFilePath)
	runtimeDirectory, _ := path.Split(constants.SocketFilePath)
	runtimeDirectoryName := path.Base(runtimeDirectory)

	tmpServiceFilePath := path.Join(os.TempDir(), daemon.serviceName())
	file, err := os.Create(tmpServiceFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.New("systemdConfig").Parse(systemdConfig)
	if err != nil {
		return err
	}
	if err := templ.Execute(
		file,
		&struct {
			BinaryPath           string
			Description          string
			Dependencies         string
			LogDir               string
			DaemonStorageDir     string
			RuntimeDirectoryName string
		}{
			BinaryPath:           path.Join(installDir, daemon.name),
			Description:          daemon.description,
			Dependencies:         strings.Join(daemon.dependencies, " "),
			RuntimeDirectoryName: runtimeDirectoryName,
			LogDir:               logDir,
			DaemonStorageDir:     daemonStorageDir,
		},
	); err != nil {
		return err
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	{
		err := sudoCommands(
			exec.Command("cp", tmpServiceFilePath, daemon.servicePath()),
			exec.Command("rm", tmpServiceFilePath),
			exec.Command("cp", binaryPath, path.Join(installDir, daemon.name)),
			exec.Command("mkdir", "-p", daemonStorageDir),
			exec.Command("mkdir", "-p", logDir),
			exec.Command("systemctl", "daemon-reload"),
			exec.Command("systemctl", "enable", daemon.serviceName()),
			exec.Command("systemctl", "start", daemon.serviceName()),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (daemon *systemDRecord) Remove() error {
	if !daemon.IsInstalled() {
		return ErrNotInstalled
	}

	if daemon.checkRunning() {
		err := sudoCommands(
			exec.Command("systemctl", "stop", daemon.serviceName()),
			exec.Command("systemctl", "disable", daemon.serviceName()),
		)
		if err != nil {
			return err
		}
	}

	logDir, _ := path.Split(constants.LogFilePath)
	DaemonStorageDir, _ := path.Split(constants.DaemonStorageFilePath)

	{
		err := sudoCommands(
			exec.Command("rm", "-f", daemon.servicePath()),
			exec.Command("rm", "-f", path.Join(installDir, daemon.name)),
			exec.Command("rm", "-rf", DaemonStorageDir),
			exec.Command("rm", "-rf", logDir),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (daemon *systemDRecord) IsInstalled() bool {
	if _, err := os.Stat(daemon.servicePath()); err == nil {
		return true
	}
	return false
}

func (daemon *systemDRecord) serviceName() string {
	return daemon.name + ".service"
}

func (daemon *systemDRecord) servicePath() string {
	return "/etc/systemd/system/" + daemon.serviceName()
}

func (daemon *systemDRecord) checkRunning() bool {
	output, err := exec.Command("sudo", "systemctl", "status", daemon.serviceName()).Output()
	if err == nil {
		if matched, err := regexp.MatchString("Active: active", string(output)); err == nil && matched {
			reg := regexp.MustCompile("Main PID: ([0-9]+)")
			data := reg.FindStringSubmatch(string(output))
			if len(data) > 1 {
				return true
			}
			return true
		}
	}

	return false
}

var systemdConfig = `[Unit]
Description={{.Description}}
Requires={{.Dependencies}}
After={{.Dependencies}}

[Service]
ExecStart={{.BinaryPath}} daemon run
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
User=root
Group=root
RestartSec=3

# Hardening
ProtectSystem=strict
ProtectKernelTunables=yes
ProtectControlGroups=yes
ProtectHome=yes
ProtectKernelModules=yes
PrivateTmp=yes
MemoryDenyWriteExecute=yes
RestrictRealtime=yes
RestrictNamespaces=yes

ReadWritePaths={{.LogDir}}
ReadWritePaths={{.DaemonStorageDir}}
RuntimeDirectory={{.RuntimeDirectoryName}}
RuntimeDirectoryMode=0775

[Install]
WantedBy=multi-user.target
`
