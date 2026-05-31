package uxHelpers

import (
	"testing"

	"github.com/zeropsio/zerops-go/types/enum"
)

// Pins the process-status decision the deploy/create spinner relies on, so the
// upcoming domain-type swap can't silently change which statuses keep polling,
// succeed, or fail.
func TestClassifyProcessStatus(t *testing.T) {
	cases := map[enum.ProcessStatusEnum]processOutcome{
		enum.ProcessStatusEnumPending:     processPolling,
		enum.ProcessStatusEnumRunning:     processPolling,
		enum.ProcessStatusEnumFinished:    processSucceeded,
		enum.ProcessStatusEnumRollbacking: processFailed,
		enum.ProcessStatusEnumCanceling:   processFailed,
		enum.ProcessStatusEnumFailed:      processFailed,
		enum.ProcessStatusEnumCanceled:    processFailed,
		enum.ProcessStatusEnum("WHATEVER_UNKNOWN"): processFailed,
	}

	for status, want := range cases {
		if got := classifyProcessStatus(status); got != want {
			t.Errorf("classifyProcessStatus(%q) = %d, want %d", status, got, want)
		}
	}
}
