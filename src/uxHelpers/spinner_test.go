package uxHelpers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/logger"
	"github.com/zeropsio/zcli/src/uxBlock"
)

// newTestBlocks builds a *uxBlock.Blocks in non-terminal mode so RunSpinners
// takes the no-op path (no Bubble Tea program is started) and all logger
// output is discarded.
func newTestBlocks() *uxBlock.Blocks {
	out := logger.NewOutputLogger(logger.OutputConfig{IsTerminal: false})
	dbg := logger.NewDebugFileLogger(logger.DebugFileConfig{})
	return uxBlock.NewBlocks(out, dbg, false, 80, 24, nil)
}

func TestProcessCheckWithSpinner_EmptyList(t *testing.T) {
	err := ProcessCheckWithSpinner(context.Background(), newTestBlocks(), nil)
	require.NoError(t, err)
}

func TestProcessCheckWithSpinner_SingleSuccess(t *testing.T) {
	var called int32
	err := ProcessCheckWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{{
			RunningMessage: "running",
			SuccessMessage: "done",
			F: func(_ context.Context, _ *Process) error {
				atomic.AddInt32(&called, 1)
				return nil
			},
		}},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestProcessCheckWithSpinner_SingleError(t *testing.T) {
	boom := errors.New("boom")
	err := ProcessCheckWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{{
			RunningMessage:      "running",
			ErrorMessageMessage: "oh no",
			F: func(_ context.Context, _ *Process) error {
				return boom
			},
		}},
	)
	require.ErrorIs(t, err, boom)
}

func TestProcessCheckWithSpinner_EmptySuccessAndErrorMessagesAreAllowed(t *testing.T) {
	// SuccessMessage / ErrorMessageMessage being empty is the documented
	// "skip the final line" sentinel — must not panic and must not change the
	// returned error.
	require.NoError(t, ProcessCheckWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{{
			F: func(_ context.Context, _ *Process) error { return nil },
		}},
	))

	boom := errors.New("boom")
	require.ErrorIs(t, ProcessCheckWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{{
			F: func(_ context.Context, _ *Process) error { return boom },
		}},
	), boom)
}

func TestProcessCheckWithSpinner_AllSucceed(t *testing.T) {
	const n = 5
	var called int32
	procs := make([]Process, n)
	for i := range procs {
		procs[i] = Process{
			RunningMessage: "running",
			SuccessMessage: "done",
			F: func(_ context.Context, _ *Process) error {
				atomic.AddInt32(&called, 1)
				return nil
			},
		}
	}
	require.NoError(t, ProcessCheckWithSpinner(context.Background(), newTestBlocks(), procs))
	require.Equal(t, int32(n), atomic.LoadInt32(&called))
}

func TestProcessCheckWithSpinner_AllRunEvenWhenOneFails(t *testing.T) {
	// Current behavior: a failure in one process does not cancel siblings;
	// every F is invoked and waited on.
	const n = 4
	var called int32
	procs := make([]Process, n)
	procs[0] = Process{F: func(_ context.Context, _ *Process) error {
		atomic.AddInt32(&called, 1)
		return errors.New("first")
	}}
	for i := 1; i < n; i++ {
		procs[i] = Process{F: func(_ context.Context, _ *Process) error {
			atomic.AddInt32(&called, 1)
			// Take a beat so the failing goroutine is likely to finish first;
			// we still assert siblings were invoked even though one returned err.
			time.Sleep(5 * time.Millisecond)
			return nil
		}}
	}
	err := ProcessCheckWithSpinner(context.Background(), newTestBlocks(), procs)
	require.Error(t, err)
	require.Equal(t, int32(n), atomic.LoadInt32(&called))
}

func TestProcessCheckWithSpinner_FirstErrorWins(t *testing.T) {
	// Current behavior: only the first error captured via sync.Once is
	// returned; subsequent errors are dropped. We can't deterministically
	// force which goroutine wins, but the returned error must be one of the
	// real ones (not a wrapper/zero value).
	errA := errors.New("err-a")
	errB := errors.New("err-b")
	err := ProcessCheckWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{
			{F: func(_ context.Context, _ *Process) error { return errA }},
			{F: func(_ context.Context, _ *Process) error { return errB }},
		},
	)
	require.Error(t, err)
	require.Truef(t, errors.Is(err, errA) || errors.Is(err, errB),
		"expected errA or errB, got %v", err)
}

func TestProcessCheckWithSpinner_SpinnerInjectedBeforeF(t *testing.T) {
	// ProcessCheckWithSpinner assigns processList[i].spinner before calling F
	// so that Process.LogView works from inside F. Lock that contract in.
	var seen *uxBlock.Spinner
	err := ProcessCheckWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{{
			F: func(_ context.Context, p *Process) error {
				seen = p.spinner
				return nil
			},
		}},
	)
	require.NoError(t, err)
	require.NotNil(t, seen, "F must receive a Process with spinner set")
}

func TestProcessCheckWithSpinner_FReceivesPointerIntoProcessList(t *testing.T) {
	// F is invoked with &processList[i]; mutating fields on the passed pointer
	// is visible to the caller after the call returns. Some callers rely on
	// this (or at least don't break if it changes), so pin current behavior.
	procs := []Process{{
		F: func(_ context.Context, p *Process) error {
			p.SuccessMessage = "mutated"
			return nil
		},
	}}
	require.NoError(t, ProcessCheckWithSpinner(context.Background(), newTestBlocks(), procs))
	require.Equal(t, "mutated", procs[0].SuccessMessage)
}

func TestProcessCheckWithSpinner_PropagatesContextToF(t *testing.T) {
	// F should observe the caller's context: cancel parent ctx and F's ctx
	// must be canceled too.
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var fErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fErr = ProcessCheckWithSpinner(
			ctx,
			newTestBlocks(),
			[]Process{{
				F: func(c context.Context, _ *Process) error {
					close(started)
					<-c.Done()
					return c.Err()
				},
			}},
		)
	}()
	<-started
	cancel()
	wg.Wait()
	require.ErrorIs(t, fErr, context.Canceled)
}

func TestProcessCheckWithSpinner_AlreadyCanceledContext(t *testing.T) {
	// If the caller hands in a context that's already canceled, F still runs
	// (ProcessCheckWithSpinner itself does not gate on ctx) and sees the
	// canceled ctx.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var observedErr error
	err := ProcessCheckWithSpinner(ctx, newTestBlocks(), []Process{{
		F: func(c context.Context, _ *Process) error {
			observedErr = c.Err()
			return nil
		},
	}})
	require.NoError(t, err)
	require.ErrorIs(t, observedErr, context.Canceled)
}
