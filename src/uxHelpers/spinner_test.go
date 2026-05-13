package uxHelpers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zcli/src/uxBlock"
)

// newTestBlocks builds a *uxBlock.Blocks in non-terminal mode so RunSpinners
// takes the no-op path (no Bubble Tea program is started).
func newTestBlocks() *uxBlock.Blocks {
	return uxBlock.NewBlocks(false, 80, 24, nil)
}

func TestRunProcessWithSpinner_Success(t *testing.T) {
	var called int32
	err := RunProcessWithSpinner(
		context.Background(),
		newTestBlocks(),
		Process{
			RunningMessage: "running",
			SuccessMessage: "done",
			F: func(_ context.Context, _ Handle) error {
				atomic.AddInt32(&called, 1)
				return nil
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestRunProcessWithSpinner_Error(t *testing.T) {
	boom := errors.New("boom")
	err := RunProcessWithSpinner(
		context.Background(),
		newTestBlocks(),
		Process{
			RunningMessage: "running",
			ErrorMessage:   "oh no",
			F: func(_ context.Context, _ Handle) error {
				return boom
			},
		},
	)
	require.ErrorIs(t, err, boom)
}

func TestRunProcessesWithSpinner_EmptyList(t *testing.T) {
	err := RunProcessesWithSpinner(context.Background(), newTestBlocks(), nil)
	require.NoError(t, err)
}

func TestRunProcessesWithSpinner_EmptyMessagesAreAllowed(t *testing.T) {
	// SuccessMessage / ErrorMessage being empty is the documented
	// "skip the final line" sentinel — must not panic and must not change
	// the returned error.
	require.NoError(t, RunProcessWithSpinner(
		context.Background(),
		newTestBlocks(),
		Process{
			F: func(_ context.Context, _ Handle) error { return nil },
		},
	))

	boom := errors.New("boom")
	require.ErrorIs(t, RunProcessWithSpinner(
		context.Background(),
		newTestBlocks(),
		Process{
			F: func(_ context.Context, _ Handle) error { return boom },
		},
	), boom)
}

func TestRunProcessesWithSpinner_AllSucceed(t *testing.T) {
	const n = 5
	var called int32
	procs := make([]Process, n)
	for i := range procs {
		procs[i] = Process{
			RunningMessage: "running",
			SuccessMessage: "done",
			F: func(_ context.Context, _ Handle) error {
				atomic.AddInt32(&called, 1)
				return nil
			},
		}
	}
	require.NoError(t, RunProcessesWithSpinner(context.Background(), newTestBlocks(), procs))
	require.Equal(t, int32(n), atomic.LoadInt32(&called))
}

func TestRunProcessesWithSpinner_FirstErrorWinsAndCancelsSiblings(t *testing.T) {
	// errgroup contract: the first non-nil error is returned, and the shared
	// context is canceled so siblings can observe the cancellation.
	boom := errors.New("boom")
	siblingSawCancel := make(chan struct{}, 1)
	err := RunProcessesWithSpinner(
		context.Background(),
		newTestBlocks(),
		[]Process{
			{F: func(_ context.Context, _ Handle) error { return boom }},
			{F: func(c context.Context, _ Handle) error {
				select {
				case <-c.Done():
					siblingSawCancel <- struct{}{}
					return c.Err()
				case <-time.After(2 * time.Second):
					return errors.New("sibling did not observe cancel")
				}
			}},
		},
	)
	require.ErrorIs(t, err, boom)
	select {
	case <-siblingSawCancel:
	default:
		t.Fatal("sibling goroutine was not signaled when first error fired")
	}
}

func TestRunProcessesWithSpinner_HandleNonNil(t *testing.T) {
	// F must receive a non-nil Handle so it can call LogView from inside the
	// process function.
	var seen Handle
	err := RunProcessWithSpinner(
		context.Background(),
		newTestBlocks(),
		Process{
			F: func(_ context.Context, h Handle) error {
				seen = h
				return nil
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, seen, "F must receive a non-nil Handle")
}

func TestRunProcessesWithSpinner_PropagatesContextToF(t *testing.T) {
	// F should observe the caller's context: cancel parent ctx and F's ctx
	// must be canceled too.
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var fErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fErr = RunProcessWithSpinner(
			ctx,
			newTestBlocks(),
			Process{
				F: func(c context.Context, _ Handle) error {
					close(started)
					<-c.Done()
					return c.Err()
				},
			},
		)
	}()
	<-started
	cancel()
	wg.Wait()
	require.ErrorIs(t, fErr, context.Canceled)
}

func TestRunProcessesWithSpinner_AlreadyCanceledContext(t *testing.T) {
	// If the caller hands in a context that's already canceled, F still runs
	// (the runner does not gate on ctx itself) and sees the canceled ctx.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var observedErr error
	err := RunProcessWithSpinner(ctx, newTestBlocks(), Process{
		F: func(c context.Context, _ Handle) error {
			observedErr = c.Err()
			return nil
		},
	})
	require.NoError(t, err)
	require.ErrorIs(t, observedErr, context.Canceled)
}
