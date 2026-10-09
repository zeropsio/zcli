package cmd

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/cliStorage"
)

// sudoServer keeps the sudo mode of each token. Approving a sudo device login switches the token that started it.
type sudoServer struct {
	lock   sync.Mutex
	inSudo map[string]bool
	status string
	// approvalIgnored leaves the session out of sudo mode after an approval.
	approvalIgnored bool
	sudoStarts      []string
	envFileReveal   []string
}

func (f *fixture) handleSudo(status string, inSudo ...string) *sudoServer {
	s := &sudoServer{inSudo: map[string]bool{}, status: status}
	for _, token := range inSudo {
		s.inSudo[token] = true
	}
	f.Mux.HandleFunc("/api/rest/public/auth/sudo", func(w http.ResponseWriter, r *http.Request) {
		s.lock.Lock()
		defer s.lock.Unlock()
		if !s.inSudo[bearer(r)] {
			writeJSON(w, http.StatusOK, map[string]any{"inSudo": false, "sudoUntil": nil})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"inSudo": true, "sudoUntil": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)})
	})
	var startedBy string
	f.Mux.HandleFunc("/api/rest/public/auth/device-login/sudo", func(w http.ResponseWriter, r *http.Request) {
		s.lock.Lock()
		defer s.lock.Unlock()
		startedBy = bearer(r)
		s.sudoStarts = append(s.sudoStarts, startedBy)
		writeJSON(w, http.StatusOK, map[string]any{
			"requestId":       "MY9Z7FZkTUyuTacCH9RoIw",
			"pollSecret":      "poll-secret",
			"userCode":        "DK0D-GF7C",
			"verificationUrl": "https://app.example.com/authorize-login?request=MY9Z7FZkTUyuTacCH9RoIw",
			"expiresIn":       300,
			"interval":        1,
		})
	})
	f.Mux.HandleFunc("/api/rest/public/auth/device-login/check", func(w http.ResponseWriter, _ *http.Request) {
		s.lock.Lock()
		defer s.lock.Unlock()
		if s.status == "APPROVED" && !s.approvalIgnored {
			s.inSudo[startedBy] = true
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": s.status, "auth": nil})
	})
	return s
}

// handleEnvFile refuses reveal=true for a token without sudo mode, like Zerops.
func (f *fixture) handleEnvFile(s *sudoServer) {
	f.HandleJSON(
		"/api/rest/public/project/"+pushProjectID,
		200,
		map[string]any{
			"id":         pushProjectID,
			"clientId":   pushClientID,
			"name":       "demo-project",
			"mode":       "LIGHT",
			"status":     "ACTIVE",
			"created":    "2024-01-01T00:00:00.000Z",
			"lastUpdate": "2024-01-01T00:00:00.000Z",
			"tagList":    []string{},
		},
	)
	f.Mux.HandleFunc("/api/rest/public/project/"+pushProjectID+"/env-file", func(w http.ResponseWriter, r *http.Request) {
		s.lock.Lock()
		defer s.lock.Unlock()
		reveal := r.URL.Query().Get("reveal")
		s.envFileReveal = append(s.envFileReveal, reveal)
		if reveal != "true" {
			writeJSON(w, http.StatusOK, map[string]any{"envFile": "SECRET=*****\n"})
			return
		}
		if !s.inSudo[bearer(r)] {
			writeJSON(w, http.StatusForbidden, apiErrorBody("notInSudoMode"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"envFile": "SECRET=shh\n"})
	})
}

func bearer(r *http.Request) string {
	return r.Header.Get("Authorization")[len("Bearer "):]
}

func (f *fixture) seedLoginSession() {
	f.seedLoginSessionExpiringIn(6 * time.Hour)
}

func (f *fixture) seedLoginSessionExpiringIn(expiresIn time.Duration) {
	f.seedData(cliStorage.Data{LoginSession: cliStorage.LoginSession{
		AccessToken:  "access-0",
		RefreshToken: "refresh-0",
		ExpiresAt:    time.Now().Add(expiresIn),
	}})
}

func TestProjectEnv_RevealEnablesSudoForLoginSession(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	sudo := f.handleSudo("APPROVED")
	f.handleEnvFile(sudo)
	f.seedLoginSession()

	res := f.Run("project", "env", "--project-id", pushProjectID, "--reveal")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Equal(t, "SECRET=shh\n", res.Stdout)
	assert.Contains(t, res.Stderr, "DK0D-GF7C")
	assert.Contains(t, res.Stderr, "Sudo mode is on.")
	assert.Equal(t, []string{"access-0"}, sudo.sudoStarts)
	assert.Equal(t, []string{"true", "true"}, sudo.envFileReveal)
}

func TestProjectEnv_RevealWithTokenWithoutSudo(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	sudo := f.handleSudo("APPROVED")
	f.handleEnvFile(sudo)
	f.SeedLogin("personal")

	res := f.Run("project", "env", "--project-id", pushProjectID, "--reveal")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "This token does not have sudo mode")
	assert.Empty(t, res.Stdout)
	assert.Empty(t, sudo.sudoStarts)
}

func TestProjectEnv_RevealWithTokenWithSudo(t *testing.T) {
	f := newFixture(t)
	sudo := f.handleSudo("APPROVED", "personal")
	f.handleEnvFile(sudo)
	f.SeedLogin("personal")

	res := f.Run("project", "env", "--project-id", pushProjectID, "--reveal")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Equal(t, "SECRET=shh\n", res.Stdout)
	assert.Empty(t, sudo.sudoStarts)
}

func TestProjectEnv_RevealInCi(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "true")
	sudo := f.handleSudo("APPROVED")
	f.handleEnvFile(sudo)
	f.seedLoginSession()

	res := f.Run("project", "env", "--project-id", pushProjectID, "--reveal")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Sudo mode cannot be enabled in CI")
	assert.Empty(t, sudo.sudoStarts)
}

func TestProjectEnv_WithoutRevealNeedsNoSudo(t *testing.T) {
	f := newFixture(t)
	sudo := f.handleSudo("APPROVED")
	f.handleEnvFile(sudo)
	f.seedLoginSession()

	res := f.Run("project", "env", "--project-id", pushProjectID)

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Equal(t, "SECRET=*****\n", res.Stdout)
	assert.Equal(t, []string{"false"}, sudo.envFileReveal)
	assert.Empty(t, sudo.sudoStarts)
}

func TestSudo_ShowsExpiryOfLoginSession(t *testing.T) {
	f := newFixture(t)
	sudo := f.handleSudo("APPROVED", "access-0")
	f.seedLoginSession()

	res := f.Run("sudo")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Sudo mode is on for 2h 0m more")
	assert.Empty(t, sudo.sudoStarts)
}

func TestSudo_TokenWithSudoHasNoExpiry(t *testing.T) {
	f := newFixture(t)
	f.handleSudo("APPROVED", "personal")
	f.SeedLogin("personal")

	res := f.Run("sudo")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Sudo mode is on.")
	assert.NotContains(t, res.Stderr, "more")
}

func TestSudo_EnablesForLoginSession(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	sudo := f.handleSudo("APPROVED")
	f.seedLoginSession()

	res := f.Run("sudo")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.Contains(t, res.Stderr, "Sudo mode is off.")
	assert.Contains(t, res.Stderr, "DK0D-GF7C")
	assert.Contains(t, res.Stderr, "Sudo mode is on for 2h 0m more")
	assert.Equal(t, []string{"access-0"}, sudo.sudoStarts)
}

func TestSudo_RefreshesSessionCloseToExpiryFirst(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	sudo := f.handleSudo("APPROVED")
	refreshes := f.handleRefresh("refresh-0")
	f.seedLoginSessionExpiringIn(10 * time.Minute)

	res := f.Run("sudo")

	require.Equalf(t, 0, res.ExitCode, "stderr=%q", res.Stderr)
	assert.EqualValues(t, 1, refreshes.Load())
	assert.Equal(t, []string{"access-1"}, sudo.sudoStarts)
	assert.Contains(t, res.Stderr, "Sudo mode is on for")
}

func TestSudo_Denied(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	f.handleSudo("DENIED")
	f.seedLoginSession()

	res := f.Run("sudo")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Sudo mode was denied in the browser.")
}

func TestProjectEnv_RevealStillRefusedAfterApproval(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	sudo := f.handleSudo("APPROVED")
	sudo.approvalIgnored = true
	f.handleEnvFile(sudo)
	f.seedLoginSession()

	res := f.Run("project", "env", "--project-id", pushProjectID, "--reveal")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Sudo mode was approved, but the session is still not in sudo mode")
	assert.NotContains(t, res.Stderr, "Sudo mode is on.")
}

func TestSudo_TokenWithoutSudo(t *testing.T) {
	f := newFixture(t)
	t.Setenv("CI", "")
	sudo := f.handleSudo("APPROVED")
	f.SeedLogin("personal")

	res := f.Run("sudo")

	require.NotEqual(t, 0, res.ExitCode)
	assert.Contains(t, res.Stderr, "Sudo mode is off.")
	assert.Contains(t, res.Stderr, "This token does not have sudo mode")
	assert.Empty(t, sudo.sudoStarts)
}

func TestRemaining(t *testing.T) {
	assert.Equal(t, "1m", remaining(10*time.Second))
	assert.Equal(t, "59m", remaining(59*time.Minute))
	assert.Equal(t, "1h 0m", remaining(time.Hour))
	assert.Equal(t, "1h 59m", remaining(119*time.Minute+20*time.Second))
}
