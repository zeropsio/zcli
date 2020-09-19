// +build darwin

package daemonInstaller

import (
	"os"
	"os/exec"
	"path"
	"regexp"
	"text/template"

	"github.com/zerops-io/zcli/src/constants"
)

const (
	installDir = "/usr/local/"
)

type darwinRecord struct {
	name         string
	description  string
	dependencies []string
}

func newDaemon(name, description string, dependencies []string) (daemon, error) {
	return &darwinRecord{
		name:         name,
		description:  description,
		dependencies: dependencies,
	}, nil
}

func (daemon *darwinRecord) Install() error {
	if daemon.IsInstalled() {
		return ErrAlreadyInstalled
	}

	cliBinaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	serviceFilePath := path.Join(os.TempDir(), daemon.name)
	file, err := os.Create(serviceFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	logDir, _ := path.Split(constants.LogFilePath)
	daemonStorageDir, _ := path.Split(constants.DaemonStorageFilePath)

	templ, err := template.New("propertyList").Parse(propertyList)
	if err != nil {
		return err
	}
	if err := templ.Execute(
		file,
		&struct {
			BinaryPath string
			Name       string
			LogFile    string
			WorkingDir string
		}{
			BinaryPath: path.Join(installDir, daemon.name),
			Name:       daemon.name,
			LogFile:    constants.LogFilePath,
			WorkingDir: daemonStorageDir,
		},
	); err != nil {
		return err
	}

	{
		err := sudoCommands(
			exec.Command("cp", serviceFilePath, daemon.servicePath()),
			exec.Command("cp", cliBinaryPath, path.Join(installDir, daemon.name)),
			exec.Command("mkdir", "-p", daemonStorageDir),
			exec.Command("mkdir", "-p", logDir),

			exec.Command("launchctl", "load", daemon.servicePath()),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (daemon *darwinRecord) Remove() error {

	if !daemon.IsInstalled() {
		return ErrNotInstalled
	}

	if daemon.checkRunning() {
		err := sudoCommands(
			exec.Command("launchctl", "unload", daemon.servicePath()),
		)
		if err != nil {
			return err
		}
	}

	daemonStorageDir, _ := path.Split(constants.DaemonStorageFilePath)

	{
		err := sudoCommands(
			exec.Command("rm", "-f", daemon.servicePath()),
			exec.Command("rm", "-f", path.Join(installDir, daemon.name)),
			exec.Command("rm", "-rf", daemonStorageDir),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (daemon *darwinRecord) IsInstalled() bool {
	if _, err := os.Stat(daemon.servicePath()); err == nil {
		return true
	}

	return false
}

func (daemon *darwinRecord) checkRunning() bool {
	output, err := exec.Command("sudo", "launchctl", "list", daemon.name).Output()
	if err == nil {
		if matched, err := regexp.MatchString(daemon.name, string(output)); err == nil && matched {
			reg := regexp.MustCompile("PID\" = ([0-9]+);")
			data := reg.FindStringSubmatch(string(output))
			if len(data) > 1 {
				return true
			}
			return true
		}
	}

	return false
}

func (daemon *darwinRecord) servicePath() string {
	return "/Library/LaunchDaemons/" + daemon.name + ".plist"
}

var propertyList = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>KeepAlive</key>
	<true/>
	<key>Label</key>
	<string>{{.Name}}</string>
	<key>ProgramArguments</key>
	<array>
	    <string>{{.BinaryPath}}</string>
	    <string>daemon</string>
	    <string>run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
    <key>WorkingDirectory</key>
    <string>{{.WorkingDir}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogFile}}</string>
    <key>StandardOutPath</key>
    <string>{{.LogFile}}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/local/sbin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
`
