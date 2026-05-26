package upgrade

import (
	"path/filepath"
	"testing"
	"time"
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
			if got := e.Fresh(); got != tc.want {
				t.Fatalf("Fresh(age=%s) = %v, want %v", tc.age, got, tc.want)
			}
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
	if err := writeCacheEntry(resp); err != nil {
		t.Fatalf("writeCacheEntry: %v", err)
	}

	got, err := loadCacheEntry()
	if err != nil {
		t.Fatalf("loadCacheEntry: %v", err)
	}
	if got == nil {
		t.Fatal("loadCacheEntry returned nil entry")
	}
	if got.Response.TagName != "v1.2.3" {
		t.Errorf("TagName = %q, want v1.2.3", got.Response.TagName)
	}
	if len(got.Response.Assets) != 1 || got.Response.Assets[0].Name != "zcli-darwin-arm64" {
		t.Errorf("Assets = %+v", got.Response.Assets)
	}
	if !got.Fresh() {
		t.Error("freshly written cache should be Fresh()")
	}
}

func TestLoadCacheEntryMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ZEROPS_CLI_DATA_FILE_PATH", filepath.Join(dir, "cli.data"))

	got, err := loadCacheEntry()
	if err != nil {
		t.Fatalf("loadCacheEntry: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil entry when cache missing, got %+v", got)
	}
}
