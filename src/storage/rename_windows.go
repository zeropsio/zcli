package storage

import (
	"errors"
	"os"
	"syscall"
	"time"
)

// errSharingViolation is ERROR_SHARING_VIOLATION, which the syscall package does not define.
const errSharingViolation syscall.Errno = 32

// Windows refuses to replace a file that another process holds open.
// constants.checkPath opens the data file without the lock at every startup, so rename retries for about 200ms.
func rename(oldPath, newPath string) error {
	const attempts = 10
	var err error
	for attempt := range attempts {
		err = os.Rename(oldPath, newPath)
		if !isHeldOpen(err) {
			return err
		}
		if attempt < attempts-1 {
			time.Sleep(time.Duration(attempt+1) * 5 * time.Millisecond)
		}
	}
	return err
}

func isHeldOpen(err error) bool {
	return errors.Is(err, syscall.ERROR_ACCESS_DENIED) || errors.Is(err, errSharingViolation)
}
