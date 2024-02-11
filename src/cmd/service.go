package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func serviceCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("service").
		Short(i18n.T(i18n.CmdService)).
		Short(i18n.T(i18n.ServiceHelp)).
		AddChildrenCmd(serviceDeleteCmd()).
		AddChildrenCmd(serviceListCmd()).
		AddChildrenCmd(serviceLogCmd()).
		AddChildrenCmd(serviceStartCmd()).
		AddChildrenCmd(serviceStopCmd()).
		AddChildrenCmd(servicePushCmd()).
		AddChildrenCmd(serviceDeployCmd())
}
