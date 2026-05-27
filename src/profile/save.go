package profile

import "os"

func (m *Manager) saveGlobal(d Data) error {
	d.StampVersion(CurrentDataVersion)
	d.Region.StampVersion(CurrentRegionVersion)
	return writeJSONAtomic(m.cfg.GlobalPath, m.cfg.GlobalMode, d)
}

// saveRepoFile persists the entire in-memory repoFile to disk. Multi-profile lives in one
// file, so any single-profile mutation (scope change, region override, token swap) rewrites
// the whole file atomically. Stamps both the file-wrapper version and every entry's own
// version so a reader can dispatch correctly even if one chain bumps without the other.
func (m *Manager) saveRepoFile() error {
	if err := m.requireRepo(); err != nil {
		return err
	}
	return writeJSONAtomic(RepoProfilePath(m.repoPath), m.cfg.GlobalMode, stampRepoFile(m.repoFile))
}

// writeRepoFile is the same as saveRepoFile but for a not-yet-loaded location. Used during
// rescueScope (global migration writes a fresh repo file) and InitRepo. Creates the parent
// .zcli/ directory if it doesn't exist yet, since this is often the first file in it.
func writeRepoFile(path string, fileMode os.FileMode, rf RepoFile) error {
	if err := ensureParentDir(path); err != nil {
		return err
	}
	return writeJSONAtomic(path, fileMode, stampRepoFile(rf))
}

// stampRepoFile returns a copy of rf with the file-wrapper version and every entry's own
// version (plus the entry's optional embedded Region) set to the current targets, all routed
// through VersionStamper. Single chokepoint so on-disk shape stays consistent no matter which
// call site writes.
func stampRepoFile(rf RepoFile) RepoFile {
	rf.StampVersion(CurrentRepoVersion)
	out := make([]RepoProfile, len(rf.Profiles))
	for i, p := range rf.Profiles {
		p.StampVersion(CurrentRepoProfileVersion)
		if p.Region != nil {
			p.Region.StampVersion(CurrentRegionVersion)
		}
		out[i] = p
	}
	rf.Profiles = out
	return rf
}
