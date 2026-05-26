package upgrade

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/constants"
)

const cacheTTL = 24 * time.Hour

type cacheEntry struct {
	FetchedAt time.Time    `json:"fetched_at"`
	Response  *apiResponse `json:"response"`
}

func (c *cacheEntry) Fresh() bool {
	return time.Since(c.FetchedAt) < cacheTTL
}

// cacheFilePath returns the path the version cache lives at, or "" if no
// writable location is available. A missing path is non-fatal — callers
// should fall back to the network.
func cacheFilePath() string {
	dataPath, _, err := constants.CliDataFilePath()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(dataPath), constants.VersionCacheFileName)
}

// loadCacheEntry reads the cache from disk. Returns (nil, nil) when no usable
// cache exists; an error indicates a malformed file the caller can ignore.
func loadCacheEntry() (*cacheEntry, error) {
	path := cacheFilePath()
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "read version cache")
	}
	entry := &cacheEntry{}
	if err := json.Unmarshal(b, entry); err != nil {
		return nil, errors.Wrap(err, "parse version cache")
	}
	if entry.Response == nil {
		return nil, nil
	}
	return entry, nil
}

func writeCacheEntry(resp *apiResponse) error {
	path := cacheFilePath()
	if path == "" {
		return nil
	}
	b, err := json.Marshal(cacheEntry{FetchedAt: time.Now(), Response: resp})
	if err != nil {
		return errors.Wrap(err, "encode version cache")
	}
	// Write to a sibling tmpfile and rename so a process exit during the
	// background refresh can't leave a half-written cache file behind.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return errors.Wrap(err, "write version cache")
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return errors.Wrap(err, "swap version cache")
	}
	return nil
}
