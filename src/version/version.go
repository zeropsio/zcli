package version

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/printer"
)

const (
	apiUrl = "https://api.github.com/repositories/269549268/releases/latest"
)

var version string
var latestResponse *apiResponse

func GetLatest(ctx context.Context) (string, error) {
	if err := fetch(ctx); err != nil {
		return "", err
	}
	return latestResponse.TagName, nil
}

func GetCurrent() string {
	return version
}

func PrintVersionCheck(ctx context.Context, out printer.Printer) {
	latestVersion, err := GetLatest(ctx)
	if err != nil {
		latestVersion = "unavailable"
	}
	latestUrl, err := GetLatestUrl(ctx)
	if err != nil {
		latestUrl = "unavailable"
	}

	if GetCurrent() == latestVersion {
		out.Printf("zcli version is up to date\n")
	} else {
		out.Printf("zcli latest available version %s\n", latestVersion)
		out.Printf("zcli latest available version download url %s\n", latestUrl)
	}
}

func PrintVersionCheckMismatch(out printer.Printer) error {
	return printMessageData(out.Writer())
}

func GetVersionCheckMismatch() (string, error) {
	b := bytes.NewBuffer(nil)
	if err := printMessageData(b); err != nil {
		return "", err
	}
	return b.String(), nil
}

func GetLatestUrl(ctx context.Context) (string, error) {
	if err := fetch(ctx); err != nil {
		return "", err
	}

	for _, asset := range latestResponse.Assets {
		if asset.Name == fmt.Sprintf("zcli-%s-%s", runtime.GOOS, runtime.GOARCH) {
			return asset.BrowserDownloadUrl, nil
		}
	}

	return "", errors.Errorf("could not find latest release for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func fetch(ctx context.Context) error {
	if latestResponse != nil {
		return nil
	}
	client := httpClient.New(ctx, httpClient.Config{HttpTimeout: time.Second * 5})
	resp, err := client.Get(ctx, apiUrl)
	if err != nil {
		return errors.Wrapf(err, "unable to get api response %s", apiUrl)
	}

	latestResponse = &apiResponse{}
	if err := json.Unmarshal(resp.Body, &latestResponse); err != nil {
		return errors.Wrap(err, "unable to read api response")
	}
	return nil
}
