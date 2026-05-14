package cmd

import (
	"context"
	"fmt"

	"github.com/zeropsio/zcli/src/entity/repository"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/kongCmd"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/wg"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types/uuid"

	"github.com/zeropsio/zcli/src/cliStorage"
)

// kongRootCLI is the kong grammar root. Its Run method renders the welcome
// when the user invokes `zcli` with no subcommand; the dispatcher handles
// "no leaf selected" by routing here.
type kongRootCLI struct {
	Env           EnvCmd           `cmd:"" help:"Display variables used by zcli."`
	Login         LoginCmd         `cmd:"" help:"Login into Zerops with generated Zerops token"`
	Logout        LogoutCmd        `cmd:"" help:"Disconnect from VPN and log out from your Zerops account"`
	Project       ProjectCmd       `cmd:"" help:"Project commands group"`
	Scope         ScopeCmd         `cmd:"" help:"Scope commands group"`
	Service       ServiceCmd       `cmd:"" help:"Service commands group"`
	ShowDebugLogs ShowDebugLogsCmd `cmd:"show-debug-logs" help:"Display zcli debug logs."`
	Support       SupportCmd       `cmd:"" help:"Display Zerops contact info."`
	Version       VersionCmd       `cmd:"" help:"Print zcli version."`
	Vpn           VpnCmd           `cmd:"" help:"VPN commands group."`
}

// Run is invoked by the dispatcher when no subcommand was provided. Prints
// the appropriate welcome (guest vs logged-in) — help is appended by the
// dispatcher after we return.
func (*kongRootCLI) Run(ctx context.Context, d *kongCmd.CmdData) error {
	auth, loggedIn := d.Auth()
	if !loggedIn {
		d.Stdout.PrintLines(i18n.T(i18n.GuestWelcome), printer.EmptyLine)
		return nil
	}

	loggedUser := "anonymous"
	if info, err := auth.RestApiClient.GetUserInfo(ctx); err != nil {
		loggedUser = err.Error()
	} else if out, err := info.Output(); err != nil {
		loggedUser = err.Error()
	} else {
		loggedUser = fmt.Sprintf("%s <%s>", out.FullName, out.Email)
	}

	if storedID, set := d.CliStorage.Data().ScopeProjectId.Get(); set {
		project, err := repository.GetProjectById(ctx, auth.RestApiClient, storedID)
		switch {
		case err != nil && errorsx.Is(err, errorsx.ErrorCode(errorCode.ProjectNotFound)):
			if _, err := d.CliStorage.Update(func(data cliStorage.Data) cliStorage.Data {
				data.ScopeProjectId = uuid.ProjectIdNull{}
				return data
			}); err != nil {
				return err
			}
		case err != nil:
			d.Stderr.PrintLines(i18n.T(i18n.ScopedProject), err.Error())
		default:
			d.Stdout.PrintLines(i18n.T(i18n.ScopedProject),
				fmt.Sprintf("%s [%s]", project.Name.String(), project.Id.Native()))
		}
	}

	vpnStatusText := i18n.T(i18n.VpnCheckingConnectionIsNotActive)
	vpnActive, err := wg.InterfaceExists()
	if err != nil {
		return err
	}
	if vpnActive {
		vpnStatusText = i18n.T(i18n.VpnCheckingConnectionIsActive)
	}
	d.Stdout.PrintLines(i18n.T(i18n.LoggedWelcome, loggedUser, vpnStatusText), printer.EmptyLine)
	return nil
}

// silence unused import lint when styles is only conditionally used
var _ = styles.InfoLine
