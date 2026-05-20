package errorsx

import (
	"fmt"

	"github.com/pkg/errors"
)

// ExitError carries a specific process exit code. RunRootCmd's error handler
// returns the code as-is without printing anything, so a command that needs a
// non-standard exit status (e.g. `upgrade --check`, which distinguishes
// up-to-date / behind / error as 0 / 1 / 2) can do its own output and signal
// the code by returning an ExitError.
type ExitError struct {
	Code int
}

func NewExitError(code int) *ExitError {
	return &ExitError{Code: code}
}

func AsExitError(err error) *ExitError {
	var exitError *ExitError
	if errors.As(err, &exitError) {
		return exitError
	}
	return nil
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}
