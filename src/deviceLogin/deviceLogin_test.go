package deviceLogin

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/errorsx"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

func TestPlain_ApprovedAfterPending(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending(), pending(), approved()})
	rec := &recorder{}

	auth, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.NoError(t, err)
	assert.Equal(t, testToken, auth.Auth.AccessToken.Native())

	initiates, checks := srv.calls()
	assert.Equal(t, 1, initiates)
	assert.Equal(t, 3, checks)
	srv.requireSecretsSent(testSecret, testSecret, testSecret)

	lines := rec.printed()
	require.Len(t, lines, 4)
	assert.Contains(t, lines[0], i18n.T(i18n.DeviceLoginConfirmCode))
	assert.Contains(t, lines[1], testCode)
	assert.Contains(t, lines[2], testUrl)
	assert.Contains(t, lines[3], i18n.T(i18n.DeviceLoginOpenLink))
}

func TestPlain_Denied(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{denied()})
	rec := &recorder{}

	_, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.ErrorIs(t, err, ErrDenied)
	require.NotNil(t, errorsx.AsUserError(err))
	assert.Equal(t, i18n.T(i18n.DeviceLoginDenied), err.Error())
}

func TestPlain_Expired(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending(), notFound()})
	rec := &recorder{}

	_, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.ErrorIs(t, err, ErrExpired)
	require.NotNil(t, errorsx.AsUserError(err))
	assert.Equal(t, i18n.T(i18n.DeviceLoginExpiredError), err.Error())

	_, checks := srv.calls()
	assert.Equal(t, 2, checks)
}

func TestPlain_TransientErrorsThenApproved(t *testing.T) {
	htmlGateway := reply{status: http.StatusBadGateway, body: "<html>bad gateway</html>"}
	srv := newFakeServer(
		t,
		[]reply{defaultInitiate()},
		[]reply{aborted(), apiErrorReply(http.StatusInternalServerError, "internalServerError"), htmlGateway, approved()},
	)
	rec := &recorder{}

	auth, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.NoError(t, err)
	assert.Equal(t, testToken, auth.Auth.AccessToken.Native())

	_, checks := srv.calls()
	assert.Equal(t, 4, checks)
	assert.Len(t, rec.debugged(), 3)
}

func TestPlain_CheckTimeoutThenApproved(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{delayed(pending(), 2*time.Second), approved()})
	rec := &recorder{}
	cfg := testConfig(t, rec, false)

	req, err := initiate(t.Context(), srv.loginFlow(), cfg)
	require.NoError(t, err)
	assert.Equal(t, 15*time.Millisecond, req.checkTimeout)

	start := time.Now()
	auth, err := Run(t.Context(), srv.api(), cfg)
	require.NoError(t, err)
	assert.Equal(t, testToken, auth.Auth.AccessToken.Native())
	assert.Less(t, time.Since(start), time.Second)

	_, checks := srv.calls()
	assert.Equal(t, 2, checks)
	require.Len(t, rec.debugged(), 1)
	assert.Contains(t, rec.debugged()[0], context.DeadlineExceeded.Error())
}

func TestPlain_ClientErrorOnCheck(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{apiErrorReply(http.StatusBadRequest, "deviceLoginAppInvalid")})
	rec := &recorder{}

	_, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.Error(t, err)
	assert.True(t, apiError.HasErrorCode(err, errorCode.DeviceLoginAppInvalid))
	assert.False(t, errors.Is(err, ErrDenied))
	assert.False(t, errors.Is(err, ErrExpired))

	_, checks := srv.calls()
	assert.Equal(t, 1, checks)
}

func TestPlain_InitiateApiError(t *testing.T) {
	srv := newFakeServer(t, []reply{apiErrorReply(http.StatusBadRequest, "deviceLoginAppInvalid")}, nil)
	rec := &recorder{}

	_, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.Error(t, err)
	assert.True(t, apiError.HasErrorCode(err, errorCode.DeviceLoginAppInvalid))

	_, checks := srv.calls()
	assert.Equal(t, 0, checks)
	assert.Empty(t, rec.printed())
}

func TestPlain_IntervalGuard(t *testing.T) {
	srv := newFakeServer(t, []reply{initiateReply(testCode, testUrl, testSecret, 0)}, []reply{pending(), approved()})
	rec := &recorder{}
	cfg := testConfig(t, rec, false)

	req, err := initiate(t.Context(), srv.loginFlow(), cfg)
	require.NoError(t, err)
	assert.Equal(t, cfg.unit, req.interval)

	start := time.Now()
	_, err = Run(t.Context(), srv.api(), cfg)
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestPlain_ContextCanceled(t *testing.T) {
	srv := newFakeServer(t, []reply{initiateReply(testCode, testUrl, testSecret, 1000)}, []reply{pending()})
	rec := &recorder{}
	ctx, cancel := context.WithCancel(t.Context())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := Run(ctx, srv.api(), testConfig(t, rec, false))
	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestPlain_ContextCanceledDuringCheck(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending()})
	rec := &recorder{}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	req, err := initiate(t.Context(), srv.loginFlow(), testConfig(t, rec, false))
	require.NoError(t, err)

	res := check(ctx, srv.loginFlow(), req)
	assert.Equal(t, pollFailed, res.outcome)
	require.ErrorIs(t, res.err, context.Canceled)
}

func TestPlain_ApprovedWithoutAuth(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{approvedWithoutAuth()})
	rec := &recorder{}

	_, err := Run(t.Context(), srv.api(), testConfig(t, rec, false))
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrDenied))
	assert.False(t, errors.Is(err, ErrExpired))
	assert.Contains(t, err.Error(), "without auth")
}

func TestPlain_OpenBrowser(t *testing.T) {
	t.Run("opened", func(t *testing.T) {
		srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{approved()})
		rec := &recorder{}
		cfg := testConfig(t, rec, false)
		var opened string
		cfg.OpenBrowser = func(url string) error {
			opened = url
			return nil
		}

		_, err := Run(t.Context(), srv.api(), cfg)
		require.NoError(t, err)
		assert.Equal(t, testUrl, opened)
		assert.Contains(t, rec.printed()[3], i18n.T(i18n.DeviceLoginBrowserOpened))
	})

	t.Run("failed", func(t *testing.T) {
		srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{approved()})
		rec := &recorder{}
		cfg := testConfig(t, rec, false)
		cfg.OpenBrowser = func(string) error {
			return errors.New("no browser")
		}

		_, err := Run(t.Context(), srv.api(), cfg)
		require.NoError(t, err)
		assert.Contains(t, rec.printed()[3], i18n.T(i18n.DeviceLoginOpenLink))
		require.Len(t, rec.debugged(), 1)
		assert.Contains(t, rec.debugged()[0], "no browser")
	})
}

func TestRunSudo_ApprovedWithoutAuth(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending(), approvedWithoutAuth()})
	rec := &recorder{}

	err := RunSudo(t.Context(), srv.api(), testConfig(t, rec, false))
	require.NoError(t, err)

	assert.Equal(t, []string{"/api/rest/public/auth/device-login/sudo"}, srv.paths())
	srv.requireSecretsSent(testSecret, testSecret)
	assert.Contains(t, rec.printed()[1], testCode)
}

func TestRunSudo_Denied(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{denied()})
	rec := &recorder{}

	err := RunSudo(t.Context(), srv.api(), testConfig(t, rec, false))
	require.ErrorIs(t, err, ErrDenied)
	assert.Equal(t, i18n.T(i18n.DeviceLoginSudoDenied), err.Error())
}

func TestRunSudo_Interactive(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{approvedWithoutAuth()})
	rec := &recorder{}
	api := srv.api()
	m := newModel(t.Context(), flow{start: starter(api.PostAuthDeviceLoginSudo), checker: api, sudo: true}, testConfig(t, rec, true))
	t.Cleanup(m.cancel)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	_, err := finalResult(t, tm)
	require.NoError(t, err)
	assert.Equal(t, []string{"/api/rest/public/auth/device-login/sudo"}, srv.paths())
}
