package upgrade

import (
	"bytes"
	"context"
	"crypto"
	_ "crypto/sha256" // register SHA-256 for selfupdate.Apply
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/httpClient"
)

const (
	checksumsName          = "checksums.txt"
	defaultDownloadTimeout = 2 * time.Minute
	defaultReleasesURL     = "https://github.com/zeropsio/zcli/releases/download/%s/%s"
	// firstSelfUpgradableTag is the earliest release whose assets include a
	// checksums.txt. Older tags can't be fetched/verified by Apply, so the
	// upgrade command refuses them and points users at install.sh instead.
	firstSelfUpgradableTag = "v1.1.0"
)

// Upgrader bundles the per-binary state and dependencies the upgrade flow
// needs: the running binary's version, the release-asset URL template, and
// the binary-swap implementation. Production callers get the real selfupdate
// path via NewUpgrader. Tests construct their own with stubs (httptest URL,
// recording apply func, custom current) instead of mutating package globals.
type Upgrader struct {
	current         string
	channel         string
	apply           func(io.Reader, selfupdate.Options) error
	releasesURL     string
	downloadTimeout time.Duration
}

// NewUpgrader builds the production Upgrader. The current version comes from
// the ldflag-stamped `version` var, but an env override (VersionEnvVar) wins
// when set - tests in other packages use that to inject without poking
// package state, mirroring how VersionApiUrlEnvVar overrides the API URL.
// The channel is read straight from the ldflag-stamped `channel` var; no
// env override since the only callers that need to vary it are same-package
// tests that construct an Upgrader literal directly.
func NewUpgrader() Upgrader {
	current := version
	if v := os.Getenv(constants.VersionEnvVar); v != "" {
		current = v
	}
	return Upgrader{
		current:         current,
		channel:         channel,
		apply:           selfupdate.Apply,
		releasesURL:     defaultReleasesURL,
		downloadTimeout: defaultDownloadTimeout,
	}
}

func (u Upgrader) Current() string { return u.current }

// WithDownloadTimeout returns a copy of u that uses d as the overall
// timeout for both checksums and binary fetch in Apply.
func (u Upgrader) WithDownloadTimeout(d time.Duration) Upgrader {
	u.downloadTimeout = d
	return u
}

type Options struct {
	// TargetVersion is the release tag to install (e.g. "v0.9.0"). Empty
	// means the latest known release.
	TargetVersion string
}

type Plan struct {
	current string
	target  string
}

func (p Plan) Current() string { return p.current }
func (p Plan) Target() string  { return p.target }

// NeedsUpgrade reports whether the upgrade command should proceed. Semantics
// differ from isUpdateAvailable (which gates the passive "newer available"
// warning): here the user explicitly asked to upgrade, so we only refuse when
// we can prove current is at or past target. A non-semver current (e.g.
// "local" dev build) against a real semver target still counts as needing the
// upgrade, since we can't prove otherwise and offering the released binary is
// the helpful default. semver.Compare ignores build metadata, so a local
// build stamped as `vX.Y.Z+N.gHASH` ties with the released `vX.Y.Z` and is
// reported as up to date.
// RequireSelfUpgradable returns an error when target predates the first
// release that shipped checksums.txt (firstSelfUpgradableTag). Apply would
// otherwise fail mid-way at the checksums fetch; this surfaces the right
// remediation up front. Non-semver targets (shouldn't happen in practice
// since target comes from a tag) pass through.
func (p Plan) RequireSelfUpgradable() error {
	if !semver.IsValid(p.target) {
		return nil
	}
	if semver.Compare(p.target, firstSelfUpgradableTag) >= 0 {
		return nil
	}
	return errors.Errorf(
		"%s predates self-upgrade support (added in %s). Install older releases via install.sh:\n  curl -fsSL https://raw.githubusercontent.com/zeropsio/zcli/main/install.sh | sh -s -- %s",
		p.target, firstSelfUpgradableTag, p.target,
	)
}

func (p Plan) NeedsUpgrade() bool {
	if !semver.IsValid(p.target) {
		return false
	}
	if !semver.IsValid(p.current) {
		return true
	}
	return semver.Compare(p.current, p.target) < 0
}

// PlanUpgrade resolves the upgrade target. Always succeeds for valid input,
// regardless of install method, so callers like `--check` can report status
// for package-managed installs too. Use RequireSelfUpdatable to enforce the
// channel restriction at the point where you actually intend to swap.
func (u Upgrader) PlanUpgrade(ctx context.Context, opts Options) (Plan, error) {
	target := opts.TargetVersion
	if target == "" {
		resp, err := fetch(ctx)
		if err != nil {
			return Plan{}, errors.Wrap(err, "resolve latest version")
		}
		target = resp.TagName
	}
	return Plan{current: u.current, target: target}, nil
}

// RequireSelfUpdatable returns an error when the running binary was installed
// through a package manager and shouldn't be replaced in place. Callers
// should run this before calling Apply.
func (u Upgrader) RequireSelfUpdatable() error {
	if method := u.Detect(); method.IsPackageManager() {
		return errors.Errorf("zcli was installed via %s; %s", method, method.Hint())
	}
	return nil
}

// Apply downloads the target binary, verifies its sha256 against the
// release's checksums.txt, and atomically swaps the running binary.
func (u Upgrader) Apply(ctx context.Context, plan Plan) error {
	asset := assetName()
	if u.downloadTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, u.downloadTimeout)
		defer cancel()
	}

	checksumsBody, err := u.httpGet(ctx, u.assetUrl(plan.target, checksumsName))
	if err != nil {
		return errors.Wrap(err, "fetch checksums.txt")
	}
	expected, err := parseChecksum(string(checksumsBody), asset)
	if err != nil {
		return err
	}

	binary, err := u.httpGet(ctx, u.assetUrl(plan.target, asset))
	if err != nil {
		return errors.Wrap(err, "download binary")
	}

	if err := u.apply(bytes.NewReader(binary), selfupdate.Options{
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

func (u Upgrader) assetUrl(tag, asset string) string {
	return fmt.Sprintf(u.releasesURL, tag, asset)
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

func (u Upgrader) httpGet(ctx context.Context, url string) ([]byte, error) {
	client := httpClient.New(ctx, httpClient.Config{HttpTimeout: u.downloadTimeout})
	resp, err := client.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s: status %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}
