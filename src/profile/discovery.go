package profile

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	// RepoProfileDir is the per-repo directory that holds a profile.json for the repo. Lives at
	// the repo root next to .git/ once a repo profile is initialised.
	RepoProfileDir = ".zcli"

	// RepoProfileFile is the per-repo profile JSON filename.
	RepoProfileFile = "profile.json"
)

// DiscoverRepoProfileDir walks up from startDir looking for the first ancestor that either
// holds a .zcli/profile.json (returned as the repo profile dir) or a .git directory (boundary -
// no profile in this repo, return ""). Walking stops at the filesystem root. Returns "" when
// no profile is found in the current repo or ancestor tree.
//
// The repo profile lookup wins over the .git/ boundary at the same level so a repo whose root
// directory holds both .git/ and .zcli/profile.json is treated as having a repo profile.
func DiscoverRepoProfileDir(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		profilePath := filepath.Join(dir, RepoProfileDir, RepoProfileFile)
		if info, statErr := os.Stat(profilePath); statErr == nil && !info.IsDir() {
			return dir, nil
		} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}

		gitDir := filepath.Join(dir, ".git")
		if info, statErr := os.Stat(gitDir); statErr == nil && info.IsDir() {
			return "", nil
		} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// RepoProfilePath returns the full path to the profile.json for the repo rooted at repoDir.
// Does not check existence - use DiscoverRepoProfileDir for that.
func RepoProfilePath(repoDir string) string {
	return filepath.Join(repoDir, RepoProfileDir, RepoProfileFile)
}

// DiscoverRepoRoot walks up from startDir looking for the first ancestor containing a .git
// directory. Used when bootstrapping a new repo profile (we want to place .zcli/ at the repo
// root, not in some arbitrary subdirectory). Returns "" when not inside a git repo at all.
func DiscoverRepoRoot(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		gitDir := filepath.Join(dir, ".git")
		if info, statErr := os.Stat(gitDir); statErr == nil && info.IsDir() {
			return dir, nil
		} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}
