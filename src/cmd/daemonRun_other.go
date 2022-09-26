//go:build !windows
// +build !windows

package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonServer"
	"github.com/zerops-io/zcli/src/vpn"
)

func createDaemonGrpcServer(vpn *vpn.Handler) (*daemonServer.Handler, error) {
	socketDir := filepath.Dir(constants.DaemonAddress)
	err := os.MkdirAll(socketDir, 0777)
	if err != nil {
		return nil, err
	}

	return daemonServer.New(daemonServer.Config{
		Socket: constants.DaemonAddress,
	}, vpn), nil
}

func prepareEnvironment() error {
	return nil
}

func run(cmd *cobra.Command, _ []string) error {
	return daemonRun(cmd.Context())
}
