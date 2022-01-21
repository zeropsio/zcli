//go:build !windows
// +build !windows

package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/grpcDaemonServer"
	"github.com/zerops-io/zcli/src/vpn"
)

func createDaemonGrpcServer(vpn *vpn.Handler) (*grpcDaemonServer.Handler, error) {
	socketDir := filepath.Dir(constants.DaemonAddress)
	err := os.MkdirAll(socketDir, 0777)
	if err != nil {
		return nil, err
	}

	return grpcDaemonServer.New(grpcDaemonServer.Config{
		Socket: constants.DaemonAddress,
	}, vpn), nil
}

func prepareEnvironment() error {
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	return daemonRun(cmd, args)
}
