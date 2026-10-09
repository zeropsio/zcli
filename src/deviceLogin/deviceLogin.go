// Package deviceLogin runs the browser approval of a public app: initiate a request, show the user code and poll until the user approves it.
// It serves both a login and the switch of an existing login session to sudo mode.
package deviceLogin

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/sdk"
)

type Checker interface {
	PostAuthDeviceLoginCheck(ctx context.Context, in body.DeviceLoginCheck) (sdk.PostAuthDeviceLoginCheckResponse, error)
}

// Api is satisfied by an unauthenticated *zeropsRestApiClient.Handler.
type Api interface {
	PostAuthDeviceLogin(ctx context.Context, in body.DeviceLoginInitiate) (sdk.PostAuthDeviceLoginResponse, error)
	Checker
}

// SudoApi is satisfied by an authorized *zeropsRestApiClient.Handler, the server binds the sudo request to the session of its token.
type SudoApi interface {
	PostAuthDeviceLoginSudo(ctx context.Context, in body.DeviceLoginInitiate) (sdk.PostAuthDeviceLoginSudoResponse, error)
	Checker
}

type flow struct {
	start   func(ctx context.Context, in body.DeviceLoginInitiate) (output.DeviceLoginInitiate, error)
	checker Checker
	// sudo is approved without auth, because the approval switches the session that started the request to sudo mode.
	sudo bool
}

type initiateResponse interface {
	Output() (output.DeviceLoginInitiate, error)
}

func starter[R initiateResponse](call func(context.Context, body.DeviceLoginInitiate) (R, error)) func(context.Context, body.DeviceLoginInitiate) (output.DeviceLoginInitiate, error) {
	return func(ctx context.Context, in body.DeviceLoginInitiate) (output.DeviceLoginInitiate, error) {
		resp, err := call(ctx, in)
		if err != nil {
			return output.DeviceLoginInitiate{}, err
		}
		return resp.Output()
	}
}

type Config struct {
	AppCode string
	// Interactive runs the Bubble Tea TUI, otherwise everything goes through UxBlocks as plain lines.
	Interactive bool
	// OpenBrowser is best effort, nil skips the attempt. On an error the user is told to open the printed link.
	OpenBrowser func(url string) error
	UxBlocks    uxBlock.UxBlocks

	// unit is the length of one server second, tests shorten it to keep polling and the countdown fast.
	unit time.Duration
}

var (
	ErrDenied  = errors.New("device login denied")
	ErrExpired = errors.New("device login expired")
)

// Run returns the AuthLogin of an approved request.
// Denied requests fail with ErrDenied, a request that the server no longer knows fails with ErrExpired in non-interactive mode
// and offers a new code in interactive mode. Ctrl+C in the TUI fails with models.ErrCtrlC.
func Run(ctx context.Context, api Api, cfg Config) (output.AuthLogin, error) {
	return run(ctx, flow{start: starter(api.PostAuthDeviceLogin), checker: api}, cfg)
}

// RunSudo switches the session of the api token to sudo mode. It fails like Run.
func RunSudo(ctx context.Context, api SudoApi, cfg Config) error {
	_, err := run(ctx, flow{start: starter(api.PostAuthDeviceLoginSudo), checker: api, sudo: true}, cfg)
	return err
}

func run(ctx context.Context, f flow, cfg Config) (output.AuthLogin, error) {
	if cfg.unit <= 0 {
		cfg.unit = time.Second
	}
	if cfg.Interactive {
		return runInteractive(ctx, f, cfg)
	}
	return runPlain(ctx, f, cfg)
}

func deniedError(f flow) error {
	if f.sudo {
		return errorsx.NewUserError(i18n.T(i18n.DeviceLoginSudoDenied), ErrDenied)
	}
	return errorsx.NewUserError(i18n.T(i18n.DeviceLoginDenied), ErrDenied)
}

func expiredError() error {
	return errorsx.NewUserError(i18n.T(i18n.DeviceLoginExpiredError), ErrExpired)
}

// openBrowser returns the line that tells the user whether the browser opened or they have to open the printed link themselves.
func openBrowser(cfg Config, url string) string {
	if cfg.OpenBrowser == nil {
		return i18n.T(i18n.DeviceLoginOpenLink)
	}
	if err := cfg.OpenBrowser(url); err != nil {
		cfg.UxBlocks.LogDebug("device login: open browser: " + err.Error())
		return i18n.T(i18n.DeviceLoginOpenLink)
	}
	return i18n.T(i18n.DeviceLoginBrowserOpened)
}
