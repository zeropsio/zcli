package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/cliStorage"
	"github.com/zeropsio/zcli/src/constants"
)

const userInfoPath = "/api/rest/public/user/info"

func (f *fixture) handleRegions() {
	f.HandleJSON("/regions", 200, map[string]any{
		"items": []map[string]any{{"name": "prg1", "isDefault": true, "address": f.Server.URL}},
	})
}

func (f *fixture) seedData(data cliStorage.Data) {
	f.t.Helper()
	data.RegionData = f.Region
	b, err := json.Marshal(data)
	require.NoError(f.t, err, "marshal seed data")
	require.NoError(f.t, os.WriteFile(f.DataPath, b, 0o600), "write seed data")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	b, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func apiErrorBody(code string) map[string]any {
	return map[string]any{"error": map[string]any{"code": code, "message": code}}
}

// handleUserInfo answers user info only for a request with the given token.
func (f *fixture) handleUserInfo(validToken string) *atomic.Int32 {
	var unauthorized atomic.Int32
	f.Mux.HandleFunc(userInfoPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+validToken {
			unauthorized.Add(1)
			writeJSON(w, http.StatusUnauthorized, apiErrorBody("unauthorized"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"email": "tester@example.com", "fullName": "Test User"})
	})
	return &unauthorized
}

func authFull(accessToken, refreshToken string) map[string]any {
	return map[string]any{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresAt":    time.Now().Add(6 * time.Hour).UTC().Format(time.RFC3339),
	}
}

// handleDeviceLogin approves or denies the request on the first check.
func (f *fixture) handleDeviceLogin(status string) *[]map[string]any {
	var lock sync.Mutex
	var checks []map[string]any
	f.HandleJSON("/api/rest/public/auth/device-login", 200, map[string]any{
		"requestId":       "MY9Z7FZkTUyuTacCH9RoIw",
		"pollSecret":      "poll-secret",
		"userCode":        "DK0D-GF7C",
		"verificationUrl": "https://app.example.com/authorize-login?request=MY9Z7FZkTUyuTacCH9RoIw",
		"expiresIn":       300,
		"interval":        1,
	})
	f.Mux.HandleFunc("/api/rest/public/auth/device-login/check", func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		lock.Lock()
		checks = append(checks, in)
		lock.Unlock()
		response := map[string]any{"status": status, "auth": nil}
		if status == "APPROVED" {
			response["auth"] = map[string]any{
				"auth": authFull("access-1", "refresh-1"),
				"user": map[string]any{"email": "tester@example.com", "fullName": "Test User"},
			}
		}
		writeJSON(w, http.StatusOK, response)
	})
	return &checks
}

func TestLogin_TokenReplacesBrowserSession(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	f.handleUserInfo("personal")
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{AccessToken: "a", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}})

	res := f.Run("login", "personal", "--region-url", f.Server.URL+"/regions")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	stored := f.LoadStorage()
	assert.Equal(t, "personal", stored.Token)
	assert.False(t, stored.LoginSession.IsSet())
}

func TestLogin_RejectsEmptyTokenAndTokenWithBrowser(t *testing.T) {
	f := newFixture(t)

	res := f.Run("login", "")
	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "The token is empty")

	res = f.Run("login", "personal", "--browser")
	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "either a token or the --browser flag")
}

func TestLogin_EnvTokenIsVerifiedAndNotStored(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	f.handleUserInfo("from-env")
	t.Setenv(constants.CliTokenEnvVar, "from-env")

	res := f.Run("login", "--region-url", f.Server.URL+"/regions")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Using the token from the ZEROPS_TOKEN env variable")
	assert.Contains(t, res.Stderr, "Test User")
	stored := f.LoadStorage()
	assert.Empty(t, stored.Token)
	assert.Equal(t, f.Server.URL, stored.RegionData.Address)
}

func TestLogin_InvalidEnvTokenFailsWithoutBrowser(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	f.handleUserInfo("other")
	checks := f.handleDeviceLogin("APPROVED")
	t.Setenv(constants.CliTokenEnvVar, "from-env")

	res := f.Run("login", "--region-url", f.Server.URL+"/regions")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Empty(t, *checks)
}

func TestLogin_BrowserIsRefusedInCI(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	t.Setenv("CI", "true")

	res := f.Run("login", "--region-url", f.Server.URL+"/regions")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Browser login is not available in CI")
}

func TestLogin_BrowserApprovedStoresSession(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	checks := f.handleDeviceLogin("APPROVED")
	t.Setenv("CI", "")
	f.seedData(cliStorage.Data{Token: "personal"})

	res := f.Run("login", "--region-url", f.Server.URL+"/regions")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "DK0D-GF7C")
	assert.Contains(t, res.Stderr, "https://app.example.com/authorize-login?request=MY9Z7FZkTUyuTacCH9RoIw")
	assert.Contains(t, res.Stderr, "Test User")
	assert.NotContains(t, res.Stderr+res.Stdout, "poll-secret")
	require.Len(t, *checks, 1)
	assert.Equal(t, "poll-secret", (*checks)[0]["pollSecret"])

	stored := f.LoadStorage()
	assert.Empty(t, stored.Token)
	assert.Equal(t, "access-1", stored.LoginSession.AccessToken)
	assert.Equal(t, "refresh-1", stored.LoginSession.RefreshToken)
}

func TestLogin_BrowserDenied(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	f.handleDeviceLogin("DENIED")
	t.Setenv("CI", "")
	f.seedData(cliStorage.Data{})

	res := f.Run("login", "--region-url", f.Server.URL+"/regions")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "denied")
	assert.False(t, f.LoadStorage().LoginSession.IsSet())
}

func TestLoggedCommand_WithoutTokenDoesNotOfferBrowserLoginWithoutTerminal(t *testing.T) {
	f := newFixture(t)
	checks := f.handleDeviceLogin("APPROVED")

	res := f.Run("project", "list")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "unauthenticated user")
	assert.Empty(t, *checks)
}

// handleRefresh rotates refresh-0 into access-1/refresh-1 once, like Zerops.
func (f *fixture) handleRefresh(valid string) *atomic.Int32 {
	var calls atomic.Int32
	f.Mux.HandleFunc("/api/rest/public/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var in struct {
			RefreshTokenId string `json:"refreshTokenId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.RefreshTokenId != valid {
			writeJSON(w, http.StatusBadRequest, apiErrorBody("refreshTokenInvalid"))
			return
		}
		writeJSON(w, http.StatusOK, authFull("access-1", "refresh-1"))
	})
	return &calls
}

func TestLoggedCommand_RefreshesRefusedSessionToken(t *testing.T) {
	f := newFixture(t)
	unauthorized := f.handleUserInfo("access-1")
	refreshes := f.handleRefresh("refresh-0")
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)}})

	res := f.Run()

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stdout+res.Stderr, "Test User")
	assert.EqualValues(t, 1, unauthorized.Load())
	assert.EqualValues(t, 1, refreshes.Load())
	stored := f.LoadStorage()
	assert.Equal(t, "access-1", stored.LoginSession.AccessToken)
	assert.Equal(t, "refresh-1", stored.LoginSession.RefreshToken)
}

func TestLoggedCommand_RefreshesAheadOfExpiry(t *testing.T) {
	f := newFixture(t)
	unauthorized := f.handleUserInfo("access-1")
	refreshes := f.handleRefresh("refresh-0")
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Minute)}})

	res := f.Run()

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Zero(t, unauthorized.Load())
	assert.EqualValues(t, 1, refreshes.Load())
}

func TestLoggedCommand_ExpiredSessionAsksForLogin(t *testing.T) {
	f := newFixture(t)
	f.handleRefresh("other")
	f.HandleJSON("/api/rest/public/user/client-list", http.StatusUnauthorized, apiErrorBody("unauthorized"))
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)}})

	res := f.Run("project", "list")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Your login session expired")
	assert.False(t, f.LoadStorage().LoginSession.IsSet())
}

func (f *fixture) handleLogout() *[]string {
	var lock sync.Mutex
	var tokens []string
	f.Mux.HandleFunc("/api/rest/public/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		tokens = append(tokens, r.Header.Get("Authorization"))
		lock.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
	})
	return &tokens
}

func TestLogout_RevokesBrowserSession(t *testing.T) {
	f := newFixture(t)
	tokens := f.handleLogout()
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)}})

	res := f.Run("logout")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Equal(t, []string{"Bearer access-0"}, *tokens)
	assert.False(t, f.LoadStorage().LoginSession.IsSet())
}

func TestLogout_KeepsPersonalTokenValid(t *testing.T) {
	f := newFixture(t)
	tokens := f.handleLogout()
	f.SeedLogin("personal")

	res := f.Run("logout")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Empty(t, *tokens)
	assert.Empty(t, f.LoadStorage().Token)
}

func TestLogout_EnvTokenOnly(t *testing.T) {
	f := newFixture(t)
	tokens := f.handleLogout()
	t.Setenv(constants.CliTokenEnvVar, "from-env")

	res := f.Run("logout")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Empty(t, *tokens)
	assert.Contains(t, res.Stderr, "still authenticated by the ZEROPS_TOKEN env variable")
}

func TestLogin_BrowserFlagWithEnvTokenWarnsAndLogsIn(t *testing.T) {
	f := newFixture(t)
	f.handleRegions()
	checks := f.handleDeviceLogin("APPROVED")
	t.Setenv("CI", "")
	t.Setenv(constants.CliTokenEnvVar, "from-env")
	f.seedData(cliStorage.Data{})

	res := f.Run("login", "--browser", "--region-url", f.Server.URL+"/regions")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "takes precedence over the browser login")
	assert.Len(t, *checks, 1)
	assert.Equal(t, "refresh-1", f.LoadStorage().LoginSession.RefreshToken)
}

func TestLoggedCommand_TokenPrecedence(t *testing.T) {
	session := cliStorage.LoginSession{AccessToken: "session", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)}
	tests := []struct {
		name     string
		env      string
		data     cliStorage.Data
		expected string
	}{
		{name: "env wins over stored token", env: "from-env", data: cliStorage.Data{Token: "personal"}, expected: "from-env"},
		{name: "env wins over session", env: "from-env", data: cliStorage.Data{LoginSession: session}, expected: "from-env"},
		{name: "stored token", data: cliStorage.Data{Token: "personal"}, expected: "personal"},
		{name: "session", data: cliStorage.Data{LoginSession: session}, expected: "session"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			unauthorized := f.handleUserInfo(tt.expected)
			t.Setenv(constants.CliTokenEnvVar, tt.env)
			f.seedData(tt.data)

			res := f.Run()

			require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
			assert.Zero(t, unauthorized.Load())
		})
	}
}

func TestRootCommand_ExpiredSessionAsksForLogin(t *testing.T) {
	f := newFixture(t)
	f.handleRefresh("other")
	f.handleUserInfo("never")
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{AccessToken: "access-0", RefreshToken: "refresh-0", ExpiresAt: time.Now().Add(time.Hour)}})

	res := f.Run()

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Your login session expired")
}

func TestLogout_IsIdempotent(t *testing.T) {
	f := newFixture(t)
	tokens := f.handleLogout()
	f.SeedLogin("personal")

	first := f.Run("logout")
	second := f.Run("logout")

	require.Equalf(t, 0, first.ExitCode, "stderr=%q", first.Stderr)
	require.Equalf(t, 0, second.ExitCode, "stderr=%q", second.Stderr)
	assert.Contains(t, second.Stderr, "You are not logged in.")
	assert.Empty(t, *tokens)
}
