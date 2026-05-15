//go:build !devel

package version

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/printer"
)

const apiUrl = "https://api.app-prg1.zerops.io/api/rest/public/zcli/version"

var version = "local"

var (
	fetchOnce      sync.Once
	latestResponse *apiResponse
	fetchErr       error
)

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
	assetName := fmt.Sprintf("zcli-%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range resp.Assets {
		if asset.Name == assetName {
			return asset.BrowserDownloadUrl, nil
		}
	}
	return "", errors.Errorf("no release asset for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func PrintVersionCheck(ctx context.Context, out printer.Printer) {
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

func IsVersionCheckMismatch(ctx context.Context) bool {
	latestVersion, err := GetLatest(ctx)
	if err != nil {
		return false
	}
	return isUpdateAvailable(GetCurrent(), latestVersion)
}

func GetVersionCheckMismatch(ctx context.Context) (string, error) {
	b := bytes.NewBuffer(nil)
	if err := printMessageData(ctx, b); err != nil {
		return "", err
	}
	return b.String(), nil
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

func fetch(ctx context.Context) (*apiResponse, error) {
	fetchOnce.Do(func() {
		client := httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Second * 5})
		resp, err := client.Get(ctx, apiUrl)
		if err != nil {
			fetchErr = errors.Wrapf(err, "version api request to %s failed", apiUrl)
			return
		}
		if resp.StatusCode != http.StatusOK {
			fetchErr = errors.Errorf("version api %s returned status %d", apiUrl, resp.StatusCode)
			return
		}
		out := &apiResponse{}
		if err := json.Unmarshal(resp.Body, out); err != nil {
			fetchErr = errors.Wrap(err, "version api response could not be decoded")
			return
		}
		latestResponse = out
	})
	return latestResponse, fetchErr
}
