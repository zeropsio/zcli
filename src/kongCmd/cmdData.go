package kongCmd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/printer"
	"github.com/zeropsio/zcli/src/uxBlock"
)

// Runner is implemented by every leaf command. The dispatcher invokes it
// after kong parse and (optional) auth resolution.
type Runner interface {
	Run(ctx context.Context, d *CmdData) error
}

// CmdData is the shared context every command receives. The authentication
// handle is unexported; commands reach it through Auth() (presence check)
// or RequireAuth() (error on absence).
type CmdData struct {
	CliStorage        *cliStorage.Handler
	UxBlocks          *uxBlock.Blocks
	Stdout            printer.Printer
	Stderr            printer.Printer
	LocalZcliYamlFile string

	auth *Auth
}

// Auth returns the active authentication handle and a presence flag. Use
// when the command behaves differently for guests but doesn't need to error
// (the no-subcommand welcome is the canonical example).
func (d *CmdData) Auth() (*Auth, bool) {
	if d.auth == nil {
		return nil, false
	}
	return d.auth, true
}

// RequireAuth returns the active Auth or an error if the user is not
// authenticated. Commands that need REST access start with this call.
func (d *CmdData) RequireAuth() (*Auth, error) {
	if d.auth == nil {
		return nil, errors.New(i18n.T(i18n.UnauthenticatedUser))
	}
	return d.auth, nil
}
