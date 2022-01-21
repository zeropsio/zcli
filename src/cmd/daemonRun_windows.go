//go:build windows
// +build windows

package cmd

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/judwhite/go-svc"
	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/grpcDaemonServer"
	"github.com/zerops-io/zcli/src/vpn"
)

func createDaemonGrpcServer(vpn *vpn.Handler) (*grpcDaemonServer.Handler, error) {
	return grpcDaemonServer.New(grpcDaemonServer.Config{
		Address: constants.DaemonAddress,
	},
		vpn,
	), nil
}

func prepareEnvironment() error {
	path, found := os.LookupEnv("PATH")
	if !found {
		return errors.New("path not found") //FIXME: i18n
	}
	path = strings.Join(append(strings.Split(path, ";"), constants.WireguardPath), ";")
	err := os.Setenv("PATH", path)
	if err != nil {
		return err
	}

	err = os.MkdirAll(constants.DaemonInstallDir, 0777)
	return err
}

type program struct {
	cmd  *cobra.Command
	args []string
	err  error

	wg   sync.WaitGroup
	quit chan error
}

func (p *program) Init(environment svc.Environment) error {
	return nil
}

func (p *program) Start() error {
	p.quit = make(chan error, 0)

	go func() {
		err := daemonRun(p.cmd, p.args)
		if err != nil {
			close(p.quit)
			p.err = err
			p.wg.Done()
			return
		}
		p.err = <-p.quit
		p.wg.Done()
	}()

	return nil
}

func (p *program) Stop() error {
	close(p.quit)
	p.wg.Wait()
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	pr := &program{cmd: cmd, args: args}

	if err := svc.Run(pr); err != nil {
		return err
	}
	return pr.err
}
