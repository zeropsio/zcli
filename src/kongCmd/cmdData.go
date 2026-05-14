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

// CmdData is the shared context every command receives. Auth is nil when
// the user has no token; commands that need REST access call RequireAuth().
type CmdData struct {
	CliStorage        *cliStorage.Handler
	UxBlocks          *uxBlock.Blocks
	Stdout            printer.Printer
	Stderr            printer.Printer
	Config            *Config
	LocalZcliYamlFile string
	Auth              *Auth
}

// RequireAuth returns the active Auth or an error if the user is not
// authenticated. Commands that need REST access start with this call.
func (d *CmdData) RequireAuth() (*Auth, error) {
	if d.Auth == nil {
		return nil, errors.New(i18n.T(i18n.UnauthenticatedUser))
	}
	return d.Auth, nil
}
