package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
	"github.com/zeropsio/zerops-go/errorCode"
)

func ExecuteCmd() {
	cmdBuilder.ExecuteRootCmd(rootCmd())
}

func rootCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("zcli").
		SetHelpTemplate(getRootTemplate()).
		SilenceError(true).
		AddChildrenCmd(loginCmd()).
		AddChildrenCmd(logoutCmd()).
		AddChildrenCmd(versionCmd()).
		AddChildrenCmd(scopeCmd()).
		AddChildrenCmd(projectCmd()).
		AddChildrenCmd(serviceCmd()).
		AddChildrenCmd(vpnCmd()).
		AddChildrenCmd(statusShowDebugLogsCmd()).
		AddChildrenCmd(servicePushCmd()).
		AddChildrenCmd(envCmd()).
		AddChildrenCmd(supportCmd()).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			cmdData.Stdout.PrintLines(
				i18n.T(i18n.GuestWelcome),
				printer.EmptyLine,
			)

			// print the default command help
			cmdData.PrintHelp()

			return nil
		}).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			var loggedUser string
			if info, err := cmdData.RestApiClient.GetUserInfo(ctx); err != nil {
				loggedUser = err.Error()
			} else {
				if infoOutput, err := info.Output(); err != nil {
					loggedUser = err.Error()
				} else {
					loggedUser = fmt.Sprintf("%s <%s>", infoOutput.FullName, infoOutput.Email)
				}
			}

			// TODO: krls - check whole block
			if cmdData.CliStorage.Data().ScopeProjectId.Filled() {
				// project scope is set
				projectId, _ := cmdData.CliStorage.Data().ScopeProjectId.Get()
				project, err := repository.GetProjectById(ctx, cmdData.RestApiClient, projectId)
				if err != nil {
					if errorsx.Is(err, errorsx.ErrorCode(errorCode.ProjectNotFound)) {
						err := cmdBuilder.ProjectScopeReset(cmdData)
						if err != nil {
							return err
						}
					} else {
						cmdData.Stderr.PrintLines(i18n.T(i18n.ScopedProject), err.Error())
					}
				} else {
					cmdData.Stdout.PrintLines(i18n.T(i18n.ScopedProject), fmt.Sprintf("%s [%s]", project.Name.String(), project.ID.Native()))
				}
			}

			var vpnStatusText string
			vpnActive, err := wg.InterfaceExists()
			if err != nil {
				return err
			}
			if vpnActive {
				vpnStatusText = i18n.T(i18n.VpnCheckingConnectionIsActive)
			} else {
				vpnStatusText = i18n.T(i18n.VpnCheckingConnectionIsNotActive)
			}

			cmdData.Stdout.PrintLines(
				i18n.T(i18n.LoggedWelcome, loggedUser, vpnStatusText),
				printer.EmptyLine,
			)

			// print the default command help
			cmdData.PrintHelp()

			return nil
		})
}

func getRootTemplate() string {
	return styles.CobraSectionStyle().SetString("Usage:").String() + `{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + styles.CobraSectionStyle().SetString("Aliases:").String() + `
{{.NameAndAliases}}{{end}}{{if .HasExample}}

` + styles.CobraSectionStyle().SetString("Examples:").String() + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

` + styles.CobraSectionStyle().SetString("Available Commands:").String() + `{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
` + styles.CobraItemNameStyle().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
` + styles.CobraItemNameStyle().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

` + styles.CobraSectionStyle().SetString("Additional Commands:").String() + `{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
` + styles.CobraItemNameStyle().SetString("{{rpad .Name .NamePadding }}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + styles.CobraSectionStyle().SetString("Flags:").String() + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + styles.CobraSectionStyle().SetString("Global Flags:").String() + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + styles.CobraSectionStyle().SetString("Additional help topics:").String() + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
` + styles.CobraItemNameStyle().SetString("{{rpad .CommandPath .CommandPathPadding}}").String() + ` {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
