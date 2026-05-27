package cmd

import (
	"context"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/output"
	"github.com/zeropsio/zcli/src/uxHelpers"
)

func projectListCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("list").
		Short(i18n.T(i18n.CmdDescProjectList)).
		StringFlag("output", "table", i18n.T(i18n.OutputFormatFlag)).
		HelpFlag(i18n.T(i18n.CmdHelpProjectList)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			outputFormat, err := output.ParseFormat(cmdData.Params.GetString("output"))
			if err != nil {
				return err
			}

			err = uxHelpers.PrintProjectList(ctx, cmdData.RestApiClient, cmdData.Stdout, outputFormat)
			if err != nil {
				return err
			}

			return nil
		})
}
