package cmdRunner

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

func TestWrap(t *testing.T) {
	RegisterTestingT(t)

	err := &execErr{
		prev:     OperationNotPermitted,
		exitCode: 1,
	}

	Expect(errors.Is(err, OperationNotPermitted)).To(BeTrue())
	Expect(errors.Is(err, CannotFindDeviceErr)).To(BeFalse())
	Expect(err.ExitCode()).To(Equal(1))

}
