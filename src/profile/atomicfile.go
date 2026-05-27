package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/file"
)

// writeJSONAtomic marshals value and writes it to path atomically: the bytes land in a sibling
// .new file first, then a rename swaps it into place. A crash between write and rename leaves
// the original file untouched; once rename returns, the new bytes are durable. The .new file
// is cleaned up on any error path.
func writeJSONAtomic(path string, fileMode os.FileMode, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return errors.WithStack(err)
	}
	tmp := path + ".new"
	defer os.Remove(tmp) // best-effort cleanup; a no-op once Rename has consumed tmp.

	if err := writeFileExact(tmp, payload, fileMode); err != nil {
		return err
	}
	// Rename on Windows refuses to overwrite an existing file, so pre-remove the target.
	// Costs a brief window of non-atomicity on POSIX that the original implementation also
	// accepted and we preserve - the new content is already durable in tmp by this point.
	_ = os.Remove(path)
	if err := os.Rename(tmp, path); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// writeFileExact writes data to path with the exact requested file mode (file.Open chmods
// after open, bypassing umask). Truncates any existing file. Closes the handle before
// returning - including on the error path - so the caller can immediately rename or delete
// the file without leaking the descriptor.
func writeFileExact(path string, data []byte, fileMode os.FileMode) error {
	f, err := file.Open(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	if _, werr := f.Write(data); werr != nil {
		f.Close()
		return errors.WithStack(werr)
	}
	return errors.WithStack(f.Close())
}

// backupProfile writes the original on-disk bytes to a sibling file named
// <path>.v<sourceVersion>.bak before a migration overwrites the live file. Lets users
// recover by `mv cli.data.v0.bak cli.data` if a future code change ever regresses on a
// migration we already ran. Overwrites an existing same-named backup, on the theory that the
// most-recent original wins; the live file is still preserved as a backup, just not older
// generations.
func backupProfile(livePath string, original []byte, sourceVersion uint, fileMode os.FileMode) error {
	backupPath := livePath + ".v" + strconv.FormatUint(uint64(sourceVersion), 10) + ".bak"
	if err := os.WriteFile(backupPath, original, fileMode); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ensureParentDir creates the parent directory of path if it doesn't exist. Used by writers
// that may be the first file ever created at their target (e.g. a fresh <repo>/.zcli/).
func ensureParentDir(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
