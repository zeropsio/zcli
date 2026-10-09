package deviceLogin

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/models"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

const testTermWidth, testTermHeight = 80, 24

func newTestModel(t *testing.T, srv *fakeServer, rec *recorder) (*model, *teatest.TestModel) {
	t.Helper()
	m := newModel(t.Context(), srv.loginFlow(), testConfig(t, rec, true))
	t.Cleanup(m.cancel)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))
	return m, tm
}

func waitForOutput(t *testing.T, tm *teatest.TestModel, want ...string) {
	t.Helper()
	teatest.WaitFor(
		t,
		tm.Output(),
		func(out []byte) bool {
			for _, w := range want {
				if !bytes.Contains(out, []byte(w)) {
					return false
				}
			}
			return true
		},
		teatest.WithDuration(5*time.Second),
	)
}

func finalResult(t *testing.T, tm *teatest.TestModel) (string, error) {
	t.Helper()
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	require.NoError(t, err)
	_, resErr := result(tm.FinalModel(t))
	return string(out), resErr
}

func TestModel_RendersCodeAndUrlThenCtrlC(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending()})
	rec := &recorder{}
	_, tm := newTestModel(t, srv, rec)

	waitForOutput(
		t,
		tm,
		i18n.T(i18n.DeviceLoginConfirmCode),
		"D K 0 D - G F 7 C",
		"╭",
		testUrl,
		i18n.T(i18n.DeviceLoginOpenLink),
		"Waiting for approval (",
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	_, err := finalResult(t, tm)
	require.ErrorIs(t, err, models.ErrCtrlC)
}

func TestModel_Approved(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending(), approved()})
	rec := &recorder{}
	m, tm := newTestModel(t, srv, rec)

	out, err := finalResult(t, tm)
	require.NoError(t, err)
	assert.Equal(t, testToken, m.auth.Auth.AccessToken.Native())
	assert.Contains(t, out, "D K 0 D - G F 7 C")
	assert.NotContains(t, out, testSecret)

	_, checks := srv.calls()
	assert.Equal(t, 2, checks)
	srv.requireSecretsSent(testSecret, testSecret)
}

func TestModel_Denied(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{denied()})
	rec := &recorder{}
	_, tm := newTestModel(t, srv, rec)

	_, err := finalResult(t, tm)
	require.ErrorIs(t, err, ErrDenied)
}

func TestModel_InitiateError(t *testing.T) {
	srv := newFakeServer(t, []reply{apiErrorReply(http.StatusBadRequest, "deviceLoginAppInvalid")}, nil)
	rec := &recorder{}
	_, tm := newTestModel(t, srv, rec)

	_, err := finalResult(t, tm)
	require.Error(t, err)
	assert.True(t, apiError.HasErrorCode(err, errorCode.DeviceLoginAppInvalid))
}

func TestModel_ExpiredThenEnterRestarts(t *testing.T) {
	srv := newFakeServer(
		t,
		[]reply{defaultInitiate(), initiateReply(testCode2, testUrl2, testSecret2, 5)},
		[]reply{notFound(), pending(), approved()},
	)
	rec := &recorder{}
	m, tm := newTestModel(t, srv, rec)

	waitForOutput(t, tm, "D K 0 D - G F 7 C", i18n.T(i18n.DeviceLoginExpired))
	_, checks := srv.calls()
	assert.Equal(t, 1, checks)

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitForOutput(t, tm, i18n.T(i18n.DeviceLoginPreviousExpired), "A B 1 2 - C D 3 4", testUrl2)

	_, err := finalResult(t, tm)
	require.NoError(t, err)
	assert.Equal(t, testToken, m.auth.Auth.AccessToken.Native())

	initiates, checks := srv.calls()
	assert.Equal(t, 2, initiates)
	assert.Equal(t, 3, checks)
	srv.requireSecretsSent(testSecret, testSecret2, testSecret2)
}

func TestModel_EnterOutsideExpiredIsIgnored(t *testing.T) {
	srv := newFakeServer(t, []reply{defaultInitiate()}, []reply{pending()})
	rec := &recorder{}
	_, tm := newTestModel(t, srv, rec)

	waitForOutput(t, tm, "Waiting for approval (")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	srv.setChecks(approved())

	_, err := finalResult(t, tm)
	require.NoError(t, err)
	initiates, _ := srv.calls()
	assert.Equal(t, 1, initiates)
}
