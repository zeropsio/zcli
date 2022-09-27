package daemonInstaller

import (
	"os/exec"

	"github.com/zeropsio/zcli/src/i18n"

	"github.com/zeropsio/zcli/src/utils/cmdRunner"
)

type Config struct {
}

type Handler struct {
	daemon daemon
}

func New(_ Config) (*Handler, error) {
	daemon, err := newDaemon("zeropsdaemon", i18n.DaemonInstallerDesc, []string{"network.target"})
	if err != nil {
		return nil, err
	}

	return &Handler{
		daemon: daemon,
	}, nil
}

func (h *Handler) Install() error {
	return h.daemon.Install()
}

func (h *Handler) Remove() error {
	return h.daemon.Remove()
}

func (h *Handler) IsInstalled() bool {
	return h.daemon.IsInstalled()
}

func sudoCommands(cmds ...*exec.Cmd) error {
	for _, cmd := range cmds {
		sudoCmd := exec.Command("sudo", cmd.Args...)
		sudoCmd.Env = cmd.Env
		sudoCmd.Stdin = cmd.Stdin
		sudoCmd.Stderr = cmd.Stderr
		sudoCmd.Stdout = cmd.Stdout
		sudoCmd.Dir = cmd.Dir

		_, err := cmdRunner.Run(sudoCmd)
		if err != nil {
			return err
		}
	}
	return nil
}
