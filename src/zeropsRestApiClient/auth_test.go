package zeropsRestApiClient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type fakeSource struct {
	lock    sync.Mutex
	token   string
	renewed string
	renewOk bool
	err     error
	refused []string
}

func (s *fakeSource) Token(context.Context) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.token, nil
}

func (s *fakeSource) Renew(_ context.Context, refused string) (string, bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.refused = append(s.refused, refused)
	if s.err != nil || !s.renewOk {
		return "", false, s.err
	}
	s.token = s.renewed
	return s.renewed, true, nil
}

// acceptServer answers 401 to every token except valid and records the token and body of each request.
func acceptServer(t *testing.T, valid string) (*httptest.Server, *[]string, *[]string) {
	t.Helper()
	var lock sync.Mutex
	var tokens, bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lock.Lock()
		tokens = append(tokens, r.Header.Get("Authorization"))
		bodies = append(bodies, string(body))
		lock.Unlock()
		if r.Header.Get("Authorization") != "Bearer "+valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(server.Close)
	return server, &tokens, &bodies
}

// post returns the response status code.
func post(t *testing.T, client *http.Client, url string, body string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, strings.NewReader(body))
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestAuthTransport_RetriesWithRenewedTokenAndBody(t *testing.T) {
	server, tokens, bodies := acceptServer(t, "fresh")
	source := &fakeSource{token: "stale", renewed: "fresh", renewOk: true}
	client := &http.Client{Transport: &authTransport{base: newTransport(), source: source}}

	status := post(t, client, server.URL, `{"a":1}`)

	require.Equal(t, http.StatusOK, status)
	require.Equal(t, []string{"Bearer stale", "Bearer fresh"}, *tokens)
	require.Equal(t, []string{`{"a":1}`, `{"a":1}`}, *bodies)
	require.Equal(t, []string{"stale"}, source.refused)
}

func TestAuthTransport_ReturnsUnauthorizedWithoutRenewal(t *testing.T) {
	server, tokens, _ := acceptServer(t, "fresh")
	client := &http.Client{Transport: &authTransport{base: newTransport(), source: StaticToken("personal")}}

	status := post(t, client, server.URL, "")

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, []string{"Bearer personal"}, *tokens)
}

func TestAuthTransport_RetriesOnlyOnce(t *testing.T) {
	server, tokens, _ := acceptServer(t, "never")
	source := &fakeSource{token: "stale", renewed: "also-refused", renewOk: true}
	client := &http.Client{Transport: &authTransport{base: newTransport(), source: source}}

	status := post(t, client, server.URL, "")

	require.Equal(t, http.StatusUnauthorized, status)
	require.Len(t, *tokens, 2)
}

func TestAuthTransport_RenewErrorIsReturned(t *testing.T) {
	server, _, _ := acceptServer(t, "fresh")
	source := &fakeSource{token: "stale", err: errors.New("network down")}
	client := &http.Client{Transport: &authTransport{base: newTransport(), source: source}}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	require.ErrorContains(t, err, "network down")
}

func TestAuthTransport_DoesNotRetryUnreadableBody(t *testing.T) {
	server, tokens, _ := acceptServer(t, "fresh")
	source := &fakeSource{token: "stale", renewed: "fresh", renewOk: true}
	client := &http.Client{Transport: &authTransport{base: newTransport(), source: source}}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, server.URL, io.NopCloser(strings.NewReader("x")))
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Len(t, *tokens, 1)
	require.Empty(t, source.refused)
}

func TestNewTokenSourceClient_SendsBearerToken(t *testing.T) {
	server, tokens, _ := acceptServer(t, "personal")

	_, err := NewAuthorizedClient("personal", server.URL).GetUserInfo(t.Context())

	require.NoError(t, err)
	require.Equal(t, []string{"Bearer personal"}, *tokens)
}
