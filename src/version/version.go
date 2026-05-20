package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/printer"
)

const defaultApiUrl = "https://api.app-prg1.zerops.io/api/rest/public/zcli/version"

// apiURL returns the version API endpoint, honoring an env override so tests
// (and mirrors) can point it elsewhere without rebuilding.
func apiURL() string {
	if u := os.Getenv(constants.VersionApiUrlEnvVar); u != "" {
		return u
	}
	return defaultApiUrl
}

var version = "local"

func GetCurrent() string {
	return version
}

func GetLatest(ctx context.Context) (string, error) {
	resp, err := fetch(ctx)
	if err != nil {
		return "", err
	}
	return resp.TagName, nil
}

func GetLatestUrl(ctx context.Context) (string, error) {
	resp, err := fetch(ctx)
	if err != nil {
		return "", err
	}
	want := assetName()
	for _, asset := range resp.Assets {
		if asset.Name == want {
			return asset.BrowserDownloadUrl, nil
		}
	}
	return "", errors.Errorf("no release asset for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func PrintVersionCheck(ctx context.Context, out printer.Printer) {
	if !semver.IsValid(GetCurrent()) {
		return
	}
	latestVersion, err := GetLatest(ctx)
	if err != nil {
		out.Printf("zcli latest version check failed\n")
		return
	}
	if !isUpdateAvailable(GetCurrent(), latestVersion) {
		out.Printf("zcli version is up to date\n")
		return
	}
	out.Printf("zcli latest available version %s\n", latestVersion)
	if latestUrl, err := GetLatestUrl(ctx); err == nil {
		out.Printf("zcli latest available version download url %s\n", latestUrl)
	}
}

// MismatchWarning returns the formatted update warning when a newer release
// is known, or "" if there is nothing to warn about. Reads only from the
// on-disk cache — never blocks on the network. The cache is populated by
// RefreshCacheIfStale running in the background.
func MismatchWarning() string {
	current := GetCurrent()
	if !semver.IsValid(current) {
		return ""
	}
	resp := loadCached()
	if resp == nil {
		return ""
	}
	if !isUpdateAvailable(current, resp.TagName) {
		return ""
	}
	return fmt.Sprintf("zcli %s is available (you have %s). %s", resp.TagName, current, Detect().Hint())
}

// RefreshCacheIfStale updates the on-disk cache when it's missing or older
// than cacheTTL. Designed for fire-and-forget background use alongside a
// synchronous MismatchWarning() call so subsequent invocations have fresh
// data. Errors are swallowed — the next run will retry.
func RefreshCacheIfStale(ctx context.Context) {
	if entry, err := loadCacheEntry(); err == nil && entry != nil && entry.Fresh() {
		return
	}
	resp, err := fetchFromNetwork(ctx)
	if err != nil {
		return
	}
	_ = writeCacheEntry(resp)
}

// loadCached returns the cached API response or nil when the cache is
// missing, malformed, or unavailable.
func loadCached() *apiResponse {
	entry, err := loadCacheEntry()
	if err != nil || entry == nil {
		return nil
	}
	return entry.Response
}

// isUpdateAvailable reports whether latest is strictly newer than current.
// Non-semver values (notably the default "local") are treated as up to date so
// dev builds and unreleased binaries don't show a false-positive warning.
func isUpdateAvailable(current, latest string) bool {
	if !semver.IsValid(current) || !semver.IsValid(latest) {
		return false
	}
	return semver.Compare(current, latest) < 0
}

// fetch returns the latest-release response, preferring the on-disk cache.
// The cache also dedupes within a single invocation: the first call writes it
// and any later call (e.g. GetLatestUrl after GetLatest) reads it back, so
// there's no need for in-process memoization.
func fetch(ctx context.Context) (*apiResponse, error) {
	if entry, err := loadCacheEntry(); err == nil && entry != nil && entry.Fresh() {
		return entry.Response, nil
	}
	resp, err := fetchFromNetwork(ctx)
	if err != nil {
		// Stale cache beats no answer at all.
		if entry, cacheErr := loadCacheEntry(); cacheErr == nil && entry != nil {
			return entry.Response, nil
		}
		return nil, err
	}
	_ = writeCacheEntry(resp)
	return resp, nil
}

func fetchFromNetwork(ctx context.Context) (*apiResponse, error) {
	url := apiURL()
	client := httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Second * 5})
	resp, err := client.Get(ctx, url)
	if err != nil {
		return nil, errors.Wrapf(err, "version api request to %s failed", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("version api %s returned status %d", url, resp.StatusCode)
	}
	out := &apiResponse{}
	if err := json.Unmarshal(resp.Body, out); err != nil {
		return nil, errors.Wrap(err, "version api response could not be decoded")
	}
	return out, nil
}
