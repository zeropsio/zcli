package version

import (
	"bytes"
	"context"
	"crypto"
	_ "crypto/sha256" // register SHA-256 for selfupdate.Apply
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/httpClient"
)

const (
	githubAssetUrl  = "https://github.com/zeropsio/zcli/releases/download/%s/%s"
	checksumsName   = "checksums.txt"
	downloadTimeout = 2 * time.Minute
)

type UpgradeOptions struct {
	// TargetVersion is the release tag to install (e.g. "v0.9.0"). Empty
	// means the latest known release.
	TargetVersion string
}

type UpgradePlan struct {
	Current string
	Target  string
}

// PlanUpgrade resolves the upgrade target and refuses package-managed
// installs. Used by --check and as the input to Upgrade.
func PlanUpgrade(ctx context.Context, opts UpgradeOptions) (*UpgradePlan, error) {
	if method := Detect(); method.IsPackageManager() {
		return nil, errors.Errorf("zcli was installed via %s; %s", method, method.Hint())
	}
	target := opts.TargetVersion
	if target == "" {
		resp, err := fetch(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "resolve latest version")
		}
		target = resp.TagName
	}
	return &UpgradePlan{Current: GetCurrent(), Target: target}, nil
}

// Upgrade downloads the target binary, verifies its sha256 against the
// release's checksums.txt, and atomically swaps the running binary.
func Upgrade(ctx context.Context, plan *UpgradePlan) error {
	asset := assetName()
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	checksumsBody, err := httpGet(ctx, assetUrl(plan.Target, checksumsName))
	if err != nil {
		return errors.Wrap(err, "fetch checksums.txt")
	}
	expected, err := parseChecksum(string(checksumsBody), asset)
	if err != nil {
		return err
	}

	binary, err := httpGet(ctx, assetUrl(plan.Target, asset))
	if err != nil {
		return errors.Wrap(err, "download binary")
	}

	if err := selfupdate.Apply(bytes.NewReader(binary), selfupdate.Options{
		Checksum: expected,
		Hash:     crypto.SHA256,
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			return errors.New("permission denied; re-run as `sudo zcli upgrade`")
		}
		return errors.Wrap(err, "apply update")
	}
	return nil
}

func assetUrl(tag, asset string) string {
	return fmt.Sprintf(githubAssetUrl, tag, asset)
}

// assetName returns the release asset filename for the current platform.
func assetName() string {
	return assetNameFor(runtime.GOOS, runtime.GOARCH)
}

// assetNameFor matches the naming in .github/workflows/release.yml:
// linux/amd64, linux/i386 (GOARCH=386), darwin/amd64, darwin/arm64,
// windows/amd64 (named "win-x64.exe").
func assetNameFor(goos, goarch string) string {
	switch {
	case goos == "windows":
		return "zcli-win-x64.exe"
	case goarch == "386":
		return fmt.Sprintf("zcli-%s-i386", goos)
	default:
		return fmt.Sprintf("zcli-%s-%s", goos, goarch)
	}
}

// parseChecksum extracts the sha256 hex for asset out of a sha256sum-style
// body (lines of "<hex>  <name>" or "<hex> *<name>").
func parseChecksum(body, asset string) ([]byte, error) {
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) != 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name == asset {
			return hex.DecodeString(fields[0])
		}
	}
	return nil, errors.Errorf("asset %s not listed in %s", asset, checksumsName)
}

func httpGet(ctx context.Context, url string) ([]byte, error) {
	client := httpClient.New(ctx, httpClient.Config{HttpTimeout: downloadTimeout})
	resp, err := client.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s: status %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}
