package cmdBuilder

import (
	"context"
	"time"

	"github.com/zeropsio/zcli/src/deviceLogin"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/loginSession"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/uxHelpers"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/sdk"
)

// WithSudo runs call once more after it enables sudo mode, if the API refused call for missing sudo mode.
// call must be safe to repeat, so it should wrap only the request that needs sudo mode.
func (d *LoggedUserCmdData) WithSudo(ctx context.Context, call func() error) error {
	err := call()
	if !apiError.HasErrorCode(err, errorCode.NotInSudoMode) {
		return err
	}
	if !d.IsLoginSession() {
		return errorsx.NewUserError(i18n.T(i18n.SudoTokenWithout), err)
	}
	if terminal.IsInteractive() {
		confirmed, promptErr := uxHelpers.YesNoPrompt(ctx, i18n.T(i18n.SudoRequiredPrompt), prompt.WithCursorPosition(1))
		if promptErr != nil {
			return promptErr
		}
		if !confirmed {
			return errorsx.NewUserError(i18n.T(i18n.SudoRequired), err)
		}
	}
	if err := d.EnableSudo(ctx); err != nil {
		return err
	}
	err = call()
	if apiError.HasErrorCode(err, errorCode.NotInSudoMode) {
		return errorsx.NewUserError(i18n.T(i18n.SudoNotApplied), err)
	}
	d.UxBlocks.PrintInfo(styles.SuccessLine(i18n.T(i18n.SudoOn)))
	return err
}

// IsLoginSession reports whether the token comes from a browser login. Only such a session can enable sudo mode, a token has it set in the GUI.
func (d *LoggedUserCmdData) IsLoginSession() bool {
	_, ok := d.tokenSource.(*loginSession.Source)
	return ok
}

// EnableSudo asks for an approval in the browser, which switches the login session to sudo mode.
func (d *LoggedUserCmdData) EnableSudo(ctx context.Context) error {
	interactive := terminal.IsInteractive()
	if !interactive && isCI() {
		return errorsx.NewUserError(i18n.T(i18n.SudoInCi), nil)
	}
	var api deviceLogin.SudoApi = d.RestApiClient
	if session, ok := d.tokenSource.(*loginSession.Source); ok {
		api = sessionSudoApi{SudoApi: d.RestApiClient, session: session}
	}
	return deviceLogin.RunSudo(ctx, api, deviceLoginConfig(d.UxBlocks, interactive))
}

// sudoRefreshMargin covers the 5 minutes of a sudo request and the 5 minute refresh margin of loginSession.Source.Token, with time to spare.
const sudoRefreshMargin = 15 * time.Minute

// sessionSudoApi refreshes the session before each sudo request.
// The server binds the request to the access token, and a refresh before the approval would delete that token.
type sessionSudoApi struct {
	deviceLogin.SudoApi
	session *loginSession.Source
}

func (a sessionSudoApi) PostAuthDeviceLoginSudo(ctx context.Context, in body.DeviceLoginInitiate) (sdk.PostAuthDeviceLoginSudoResponse, error) {
	a.session.RefreshAhead(ctx, sudoRefreshMargin)
	return a.SudoApi.PostAuthDeviceLoginSudo(ctx, in)
}
