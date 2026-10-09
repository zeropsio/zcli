package deviceLogin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/uxBlock/mocks"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
)

const (
	testCode      = "DK0D-GF7C"
	testCode2     = "AB12-CD34"
	testUrl       = "https://tatami.devel.zerops.dev/authorize-login?request=MY9Z7FZkTUyuTacCH9RoIw"
	testUrl2      = "https://tatami.devel.zerops.dev/authorize-login?request=Y2JfkQbSRZOfHe5lGmUJoQ"
	testRequestId = "MY9Z7FZkTUyuTacCH9RoIw"
	testSecret    = "poll-secret-one"
	testSecret2   = "poll-secret-two"
	testToken     = "access-token"
)

// reply is one scripted HTTP answer. abort drops the connection, which the SDK reports as a transport error, delay holds the answer back.
type reply struct {
	status int
	body   string
	abort  bool
	delay  time.Duration
}

func delayed(rep reply, delay time.Duration) reply {
	rep.delay = delay
	return rep
}

func initiateReply(code string, url string, secret string, interval int) reply {
	body := fmt.Sprintf(
		`{"requestId":%q,"pollSecret":%q,"userCode":%q,"verificationUrl":%q,"expiresIn":%d,"interval":%d}`,
		testRequestId, secret, code, url, 300, interval,
	)
	return reply{status: http.StatusOK, body: body}
}

func defaultInitiate() reply {
	return initiateReply(testCode, testUrl, testSecret, 5)
}

func pending() reply {
	return reply{status: http.StatusOK, body: `{"status":"PENDING","auth":null}`}
}

func approved() reply {
	return reply{status: http.StatusOK, body: fmt.Sprintf(`{"status":"APPROVED","auth":{"user":null,"auth":{"accessToken":%q,"refreshToken":"refresh-token"}}}`, testToken)}
}

func approvedWithoutAuth() reply {
	return reply{status: http.StatusOK, body: `{"status":"APPROVED","auth":null}`}
}

func denied() reply {
	return reply{status: http.StatusOK, body: `{"status":"DENIED","auth":null}`}
}

func apiErrorReply(status int, code string) reply {
	return reply{status: status, body: fmt.Sprintf(`{"error":{"code":%q,"message":"test error"}}`, code)}
}

func notFound() reply {
	return apiErrorReply(http.StatusNotFound, "deviceLoginNotFound")
}

func aborted() reply {
	return reply{abort: true}
}

// fakeServer answers the device login endpoints from scripted queues. The last reply of a queue repeats.
type fakeServer struct {
	t   *testing.T
	srv *httptest.Server

	mu            sync.Mutex
	initiates     []reply
	checks        []reply
	initiateCalls int
	checkCalls    int
	// initiatePaths records the endpoint of every initiate, the login and the sudo flow share the scripted replies.
	initiatePaths []string
	// sentSecrets records the pollSecret of every check so tests can compare it without printing it.
	sentSecrets []string
}

func newFakeServer(t *testing.T, initiates []reply, checks []reply) *fakeServer {
	t.Helper()
	f := &fakeServer{t: t, initiates: initiates, checks: checks}
	f.srv = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeServer) api() sdk.Handler {
	return sdk.New(sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(f.srv.URL)), f.srv.Client())
}

func (f *fakeServer) loginFlow() flow {
	api := f.api()
	return flow{start: starter(api.PostAuthDeviceLogin), checker: api}
}

func (f *fakeServer) handle(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	var rep reply
	switch r.URL.Path {
	case "/api/rest/public/auth/device-login", "/api/rest/public/auth/device-login/sudo":
		f.initiateCalls++
		f.initiatePaths = append(f.initiatePaths, r.URL.Path)
		rep = next(&f.initiates)
	case "/api/rest/public/auth/device-login/check":
		f.checkCalls++
		var in struct {
			RequestId  string `json:"requestId"`
			PollSecret string `json:"pollSecret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			f.mu.Unlock()
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.sentSecrets = append(f.sentSecrets, in.PollSecret)
		rep = next(&f.checks)
	default:
		f.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	f.mu.Unlock()

	if rep.delay > 0 {
		select {
		case <-time.After(rep.delay):
		case <-r.Context().Done():
			return
		}
	}
	if rep.abort {
		panic(http.ErrAbortHandler)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rep.status)
	_, _ = w.Write([]byte(rep.body))
}

func next(queue *[]reply) reply {
	if len(*queue) == 0 {
		return reply{status: http.StatusInternalServerError, body: `{"error":{"code":"testQueueEmpty","message":"no reply scripted"}}`}
	}
	rep := (*queue)[0]
	if len(*queue) > 1 {
		*queue = (*queue)[1:]
	}
	return rep
}

func (f *fakeServer) setChecks(checks ...reply) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.checks = checks
}

func (f *fakeServer) calls() (initiates int, checks int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.initiateCalls, f.checkCalls
}

func (f *fakeServer) paths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.initiatePaths...)
}

func (f *fakeServer) requireSecretsSent(secrets ...string) {
	f.t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	require.Len(f.t, f.sentSecrets, len(secrets), "unexpected number of checks")
	for i, secret := range secrets {
		//nolint:testifylint // Why: a failing Equal would print the pollSecret
		require.True(f.t, f.sentSecrets[i] == secret, "check %d sent a different pollSecret", i)
	}
}

// recorder captures what the flow prints through UxBlocks.
type recorder struct {
	mu    sync.Mutex
	lines []string
	debug []string
}

func (r *recorder) uxBlocks(t *testing.T) *mocks.MockUxBlocks {
	t.Helper()
	ux := mocks.NewMockUxBlocks(gomock.NewController(t))
	ux.EXPECT().PrintInfoText(gomock.Any()).AnyTimes().Do(func(line string) {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.lines = append(r.lines, line)
	})
	ux.EXPECT().PrintInfo(gomock.Any()).AnyTimes().Do(func(line styles.Line) {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.lines = append(r.lines, line.String())
	})
	ux.EXPECT().LogDebug(gomock.Any()).AnyTimes().Do(func(message string) {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.debug = append(r.debug, message)
	})
	return ux
}

func (r *recorder) printed() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lines...)
}

func (r *recorder) debugged() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.debug...)
}

func testConfig(t *testing.T, rec *recorder, interactive bool) Config {
	t.Helper()
	return Config{
		AppCode:     "zcli",
		Interactive: interactive,
		UxBlocks:    rec.uxBlocks(t),
		unit:        time.Millisecond,
	}
}
