package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/secrets"
)

// TestLoadV0File_DrainsTokenAndVpnKeys verifies that a legacy on-disk file (no "version" key,
// PascalCase fields, plaintext token + VPN keys inline) loads as v0, drains its sensitive
// fields into the secrets backend, and gets rewritten on disk in the current snake_case shape
// with the sensitive fields gone.
func TestLoadV0File_DrainsTokenAndVpnKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	v0 := []byte(`{
		"Token":"legacy-token",
		"RegionData":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"},
		"ProjectVpnKeyRegistry":{"AAAAAAAAAAAAAAAAAAAAAA":{"Key":"wgkey-payload","ProjectId":"AAAAAAAAAAAAAAAAAAAAAA","CreatedAt":"2026-01-01T00:00:00Z"}}
	}`)
	require.NoError(t, os.WriteFile(path, v0, 0o600))

	store := secrets.NewMemory()
	m, err := NewManager(Config{
		ProfilePath: path,
		ProfileMode: 0o600,
		Secrets:     store,
	})
	require.NoError(t, err)

	// Non-sensitive state: Region survived, version stamped.
	d := m.Data()
	require.NotNil(t, d.Version)
	require.Equal(t, CurrentDataVersion, *d.Version)
	require.Equal(t, "prg1", d.Region.Name)

	// Token drained into secrets backend.
	tok, ok, err := m.Token()
	require.NoError(t, err)
	require.True(t, ok, "token should be present in secrets after v0 drain")
	require.Equal(t, "legacy-token", tok)

	// Token + one VPN key drained into secrets backend.
	keys, err := store.List(secrets.Prefix(secrets.GlobalScope))
	require.NoError(t, err)
	require.Len(t, keys, 2, "expected token + one vpn key in secrets after drain")

	// On-disk file is now in v1 shape: snake_case, version=1, no plaintext secrets.
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.EqualValues(t, CurrentDataVersion, onDisk["version"])
	require.NotContains(t, onDisk, "Token", "legacy plaintext token must be gone from disk")
	require.NotContains(t, onDisk, "token", "token must never appear in profile JSON")
	require.NotContains(t, onDisk, "ProjectVpnKeyRegistry", "legacy plaintext vpn keys must be gone from disk")
	require.NotContains(t, onDisk, "project_vpn_key_registry", "vpn keys must never appear in profile JSON")
	require.Equal(t, "prg1", onDisk["region"].(map[string]any)["name"])
}

// TestManager_SetTokenAndClear covers the round-trip through the secrets backend plus the
// Clear() sweep that wipes both the profile JSON and every global-scope secret on logout.
func TestManager_SetTokenAndClear(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	store := secrets.NewMemory()
	m, err := NewManager(Config{
		ProfilePath: path,
		ProfileMode: 0o600,
		Secrets:     store,
	})
	require.NoError(t, err)

	require.NoError(t, m.SetToken("fresh-token"))
	tok, ok, err := m.Token()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "fresh-token", tok)

	_, err = m.Clear()
	require.NoError(t, err)

	_, ok, err = m.Token()
	require.NoError(t, err)
	require.False(t, ok, "token should be cleared from secrets after Clear")

	keys, err := store.List(secrets.Prefix(secrets.GlobalScope))
	require.NoError(t, err)
	require.Empty(t, keys, "all global secrets should be removed by Clear")
}
