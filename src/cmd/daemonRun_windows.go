//go:build windows
// +build windows

package cmd

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/judwhite/go-svc"
	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonServer"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/vpn"
)

func createDaemonGrpcServer(vpn *vpn.Handler) (*daemonServer.Handler, error) {
	return daemonServer.New(daemonServer.Config{
		Address: constants.DaemonAddress,
	},
		vpn,
	), nil
}

func prepareEnvironment() error {
	path, found := os.LookupEnv("PATH")
	if !found {
		return errors.New(i18n.PathNotFound)
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

	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context
}

func (p *program) Init(environment svc.Environment) error {
	return nil
}

func (p *program) Start() error {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.cancel()
		p.err = daemonRun(p.ctx)
	}()
	return nil
}

func (p *program) Stop() error {
	p.cancel()
	p.wg.Wait()
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	pr := &program{
		cmd:  cmd,
		args: args,
	}
	pr.ctx, pr.cancel = context.WithCancel(cmd.Context())
	if err := svc.Run(pr); err != nil {
		return err
	}
	return pr.err
}
