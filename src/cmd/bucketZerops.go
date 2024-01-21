package cmd

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func bucketZeropsCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("zerops").
		Short(i18n.T(i18n.CmdBucketZerops)).
		AddChildrenCmd(bucketZeropsCreateCmd()).
		AddChildrenCmd(bucketZeropsDeleteCmd())
}
