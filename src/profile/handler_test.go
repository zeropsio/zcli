package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/storage"
)

// TestLoadV0File verifies that a legacy on-disk file (no "version" key, PascalCase fields as
// emitted by default json encoding before versioning was added) loads as v0, migrates in-memory
// to CurrentDataVersion, and gets rewritten in the current snake_case shape on the next save.
func TestLoadV0File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	v0 := []byte(`{"Token":"legacy-token","RegionData":{"name":"prg1","isDefault":true,"address":"api.app-prg1.zerops.io"}}`)
	require.NoError(t, os.WriteFile(path, v0, 0o600))

	h, err := NewHandler(storage.Config{FilePath: path, FileMode: 0o600})
	require.NoError(t, err)

	d := h.Data()
	require.NotNil(t, d.Version, "v0 file should be migrated and have version set")
	require.Equal(t, CurrentDataVersion, *d.Version)
	require.Equal(t, "legacy-token", d.Token)
	require.Equal(t, "prg1", d.RegionData.Name)

	_, err = h.Update(func(d Data) Data { return d })
	require.NoError(t, err)

	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var onDisk map[string]any
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	require.EqualValues(t, CurrentDataVersion, onDisk["version"])
	require.Equal(t, "legacy-token", onDisk["token"], "fields should be rewritten with snake_case keys")
	require.NotContains(t, onDisk, "Token", "legacy PascalCase key should be gone after re-save")
}
