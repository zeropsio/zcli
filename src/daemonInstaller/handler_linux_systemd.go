// +build linux

package daemonInstaller

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/i18n"
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

	_, err := exec.LookPath("wg")
	if err != nil {
		return errors.New(i18n.DaemonInstallWireguardNotFound)
	}

	tmpServiceFilePath := path.Join(os.TempDir(), daemon.serviceName())
	file, err := os.Create(tmpServiceFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, err := template.New("systemdConfig").Parse(systemdConfig)
	if err != nil {
		return err
	}

	// create read writes paths
	logDir := path.Dir(constants.LogFilePath)
	daemonStorageDir := path.Dir(constants.DaemonStorageFilePath)
	readWritePaths := []string{
		logDir,
		daemonStorageDir,
	}

	dnsManagement, err := dns.DetectDns()
	if err != nil {
		return err
	}
	if dnsManagement == dns.LocalDnsManagementResolveConf {
		dir := path.Dir(constants.ResolvconfOrderFilePath)
		readWritePaths = append(readWritePaths, dir)
		readWritePaths = append(readWritePaths, "/run/resolvconf/")
	}
	if dnsManagement == dns.LocalDnsManagementFile {
		dir := path.Dir(constants.ResolvFilePath)
		readWritePaths = append(readWritePaths, dir)
	}

	runtimeDirectoryName := path.Base(path.Dir(constants.SocketFilePath))

	if err := tmpl.Execute(
		file,
		&struct {
			BinaryPath           string
			Description          string
			Dependencies         string
			LogDir               string
			RuntimeDirectoryName string
			ReadWritePaths       []string
		}{
			BinaryPath:           path.Join(constants.DaemonInstallDir, daemon.name),
			Description:          daemon.description,
			Dependencies:         strings.Join(daemon.dependencies, " "),
			RuntimeDirectoryName: runtimeDirectoryName,
			ReadWritePaths:       readWritePaths,
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
			exec.Command("cp", binaryPath, path.Join(constants.DaemonInstallDir, daemon.name)),
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

	logDir := path.Dir(constants.LogFilePath)
	DaemonStorageDir := path.Dir(constants.DaemonStorageFilePath)

	{
		err := sudoCommands(
			exec.Command("rm", "-f", daemon.servicePath()),
			exec.Command("rm", "-f", path.Join(constants.DaemonInstallDir, daemon.name)),
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

{{ range .ReadWritePaths -}}
ReadWritePaths={{ . }}
{{ end -}}
RuntimeDirectory={{.RuntimeDirectoryName}}
RuntimeDirectoryMode=0775

[Install]
WantedBy=multi-user.target
`
