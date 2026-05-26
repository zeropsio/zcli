package upgrade

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheEntryFresh(t *testing.T) {
	cases := []struct {
		name string
		age  time.Duration
		want bool
	}{
		{"just now", 0, true},
		{"one hour ago", time.Hour, true},
		{"under ttl", cacheTTL - time.Minute, true},
		{"at ttl boundary", cacheTTL + time.Minute, false},
		{"a week ago", 7 * 24 * time.Hour, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &cacheEntry{FetchedAt: time.Now().Add(-tc.age)}
			assert.Equalf(t, tc.want, e.Fresh(), "Fresh(age=%s)", tc.age)
		})
	}
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ZEROPS_CLI_DATA_FILE_PATH", filepath.Join(dir, "cli.data"))

	resp := &apiResponse{
		TagName: "v1.2.3",
		Assets: []apiAsset{
			{Name: "zcli-darwin-arm64", BrowserDownloadUrl: "https://example.com/zcli"},
		},
	}
	require.NoError(t, writeCacheEntry(resp))

	got, err := loadCacheEntry()
	require.NoError(t, err)
	require.NotNil(t, got, "loadCacheEntry returned nil entry")
	assert.Equal(t, "v1.2.3", got.Response.TagName)
	require.Len(t, got.Response.Assets, 1)
	assert.Equal(t, "zcli-darwin-arm64", got.Response.Assets[0].Name)
	assert.True(t, got.Fresh(), "freshly written cache should be Fresh()")
}

func TestLoadCacheEntryMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ZEROPS_CLI_DATA_FILE_PATH", filepath.Join(dir, "cli.data"))

	got, err := loadCacheEntry()
	require.NoError(t, err)
	assert.Nil(t, got, "expected nil entry when cache missing")
}
