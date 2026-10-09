package loginSession

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/storage"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
)

// refreshServer rotates the token pair like Zerops does: a refresh token works once.
type refreshServer struct {
	server  *httptest.Server
	calls   atomic.Int32
	lock    sync.Mutex
	valid   map[string]bool
	next    int
	expires time.Duration
	failAll bool
}

func newRefreshServer(t *testing.T, validRefreshToken string) *refreshServer {
	t.Helper()
	s := &refreshServer{valid: map[string]bool{validRefreshToken: true}, expires: 6 * time.Hour}
	s.server = httptest.NewServer(http.HandlerFunc(s.handle))
	t.Cleanup(s.server.Close)
	return s
}

func (s *refreshServer) handle(w http.ResponseWriter, r *http.Request) {
	s.calls.Add(1)
	w.Header().Set("Content-Type", "application/json")
	if s.failAll {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"internalServerError","message":"boom"}}`))
		return
	}
	var in struct {
		RefreshTokenId string `json:"refreshTokenId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.valid[in.RefreshTokenId] {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"refreshTokenInvalid","message":"invalid"}}`))
		return
	}
	delete(s.valid, in.RefreshTokenId)
	s.next++
	refresh := fmt.Sprintf("refresh-%d", s.next)
	s.valid[refresh] = true
	_, _ = fmt.Fprintf(
		w,
		`{"accessToken":"access-%d","refreshToken":%q,"expiresAt":%q}`,
		s.next, refresh, time.Now().Add(s.expires).UTC().Format(time.RFC3339),
	)
}

func newStorage(t *testing.T, path string) *cliStorage.Handler {
	t.Helper()
	s, err := storage.New[cliStorage.Data](storage.Config{FilePath: path, FileMode: 0o600})
	require.NoError(t, err)
	return &cliStorage.Handler{Handler: s}
}

func seed(t *testing.T, path string, session cliStorage.LoginSession) {
	t.Helper()
	_, err := newStorage(t, path).Update(func(data cliStorage.Data) cliStorage.Data {
		data.LoginSession = session
		return data
	})
	require.NoError(t, err)
}

func newSource(t *testing.T, path string, server *refreshServer) *Source {
	t.Helper()
	return New(newStorage(t, path), zeropsRestApiClient.NewClient(server.server.URL))
}

func TestToken_FreshTokenIsNotRefreshed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)})

	token, err := newSource(t, path, server).Token(t.Context())

	require.NoError(t, err)
	require.Equal(t, "access-0", token)
	require.Zero(t, server.calls.Load())
}

func TestToken_RefreshesAheadOfExpiryAndStoresThePair(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Minute)})

	token, err := newSource(t, path, server).Token(t.Context())

	require.NoError(t, err)
	require.Equal(t, "access-1", token)
	stored := newStorage(t, path).Data().LoginSession
	require.Equal(t, "access-1", stored.AccessToken)
	require.Equal(t, "refresh-1", stored.RefreshToken)
	require.True(t, stored.ExpiresAt.After(time.Now().Add(5*time.Hour)))
}

func TestRefreshAhead_RefreshesWithinTheGivenMargin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(10 * time.Minute)})
	source := newSource(t, path, server)

	token, err := source.Token(t.Context())
	require.NoError(t, err)
	require.Equal(t, "access-0", token)

	source.RefreshAhead(t.Context(), 15*time.Minute)
	token, err = source.Token(t.Context())
	require.NoError(t, err)
	require.Equal(t, "access-1", token)
	require.EqualValues(t, 1, server.calls.Load())
}

func TestRenew_AdoptsTokenThatAnotherProcessRefreshed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)})
	first := newSource(t, path, server)
	second := newSource(t, path, server)

	renewed, ok, err := first.Renew(t.Context(), "access-0")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "access-1", renewed)

	renewed, ok, err = second.Renew(t.Context(), "access-0")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "access-1", renewed)
	require.EqualValues(t, 1, server.calls.Load())
}

func TestRenew_ConcurrentProcessesRefreshOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)})

	var wg sync.WaitGroup
	results := make([]string, 4)
	for i := range results {
		source := newSource(t, path, server)
		wg.Go(func() {
			renewed, ok, err := source.Renew(t.Context(), "access-0")
			require.NoError(t, err)
			require.True(t, ok)
			results[i] = renewed
		})
	}
	wg.Wait()

	require.EqualValues(t, 1, server.calls.Load())
	for _, renewed := range results {
		require.Equal(t, "access-1", renewed)
	}
}

func TestRenew_RejectedRefreshTokenClearsTheSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "other")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)})
	source := newSource(t, path, server)

	_, ok, err := source.Renew(t.Context(), "access-0")

	require.NoError(t, err)
	require.False(t, ok)
	require.True(t, source.Expired())
	require.False(t, newStorage(t, path).Data().LoginSession.IsSet())
}

func TestRenew_ServerErrorKeepsTheSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	server.failAll = true
	session := cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Second)}
	seed(t, path, session)
	source := newSource(t, path, server)

	_, ok, err := source.Renew(t.Context(), "access-0")

	require.Error(t, err)
	require.False(t, ok)
	require.False(t, source.Expired())
	require.Equal(t, session.RefreshToken, newStorage(t, path).Data().LoginSession.RefreshToken)
}

func TestRenew_LoggedOutElsewhereIsExpired(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cli.data")
	server := newRefreshServer(t, "refresh-0")
	seed(t, path, cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)})
	source := newSource(t, path, server)
	seed(t, path, cliStorage.LoginSession{})

	_, ok, err := source.Renew(t.Context(), "access-0")

	require.NoError(t, err)
	require.False(t, ok)
	require.True(t, source.Expired())
	require.Zero(t, server.calls.Load())
}
