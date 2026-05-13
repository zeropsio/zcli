package uxBlock

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestBlocks builds a *Blocks in non-terminal mode so RunSpinners
// takes the no-op path (no Bubble Tea program is started). Lets us
// exercise the WithSpinner outcome plumbing without a TTY.
func newTestBlocks() *Blocks {
	return NewBlocks(false, 80, 24, nil)
}

func TestWithSpinner_SuccessReturnsNil(t *testing.T) {
	var called int32
	err := newTestBlocks().WithSpinner(
		context.Background(),
		SpinnerSpec{Message: "running", SuccessMessage: "done"},
		func(h SpinnerHandle) error {
			require.NotNil(t, h, "WithSpinner must pass a non-nil handle")
			atomic.AddInt32(&called, 1)
			return nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestWithSpinner_ErrorIsReturnedUnchanged(t *testing.T) {
	// Mirrors the showcase's "error path with custom ErrorMessage".
	boom := errors.New("oh no")
	err := newTestBlocks().WithSpinner(
		context.Background(),
		SpinnerSpec{Message: "running risky thing", ErrorMessage: "risky thing failed"},
		func(_ SpinnerHandle) error { return boom },
	)
	require.ErrorIs(t, err, boom)
}

func TestWithSpinner_ErrorWithoutCustomMessageFallsBackToErrError(t *testing.T) {
	// Mirrors the showcase's "error path without ErrorMessage". The
	// outcome line uses err.Error(), but the return value is still the
	// raw error.
	boom := errors.New("something broke at step 42")
	err := newTestBlocks().WithSpinner(
		context.Background(),
		SpinnerSpec{Message: "trying something"},
		func(_ SpinnerHandle) error { return boom },
	)
	require.ErrorIs(t, err, boom)
}

func TestWithSpinner_SetMessageDoesNotPanic(t *testing.T) {
	// Mirrors the showcase's success-with-SetMessage flow. In
	// non-terminal mode SetMessage is a no-op (the send function is a
	// no-op), but the call must remain safe.
	err := newTestBlocks().WithSpinner(
		context.Background(),
		SpinnerSpec{Message: "preparing", SuccessMessage: "deployed"},
		func(h SpinnerHandle) error {
			h.SetMessage("uploading package")
			h.SetMessage("running deploy")
			return nil
		},
	)
	require.NoError(t, err)
}

func TestWithSpinner_LogViewReturnsStdoutWhenNotTerminal(t *testing.T) {
	// The handle's LogView should fall back to os.Stdout when there's
	// no terminal, so writes never panic.
	var got any
	err := newTestBlocks().WithSpinner(
		context.Background(),
		SpinnerSpec{Message: "running"},
		func(h SpinnerHandle) error {
			got = h.LogView()
			return nil
		},
	)
	require.NoError(t, err)
	require.Same(t, os.Stdout, got, "LogView() should return os.Stdout in non-terminal mode")
}

func TestWithSpinner_PropagatesContextToClosure(t *testing.T) {
	// Cancelling the parent context must be observable inside the
	// closure (the spinner runner does not gate on ctx itself; only
	// the closure does).
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var fErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fErr = newTestBlocks().WithSpinner(
			ctx,
			SpinnerSpec{Message: "running"},
			func(_ SpinnerHandle) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			},
		)
	}()
	<-started
	cancel()
	wg.Wait()
	require.ErrorIs(t, fErr, context.Canceled)
}

func TestWithSpinner_AlreadyCanceledContext(t *testing.T) {
	// An already-canceled context still runs the closure (WithSpinner
	// does not gate on ctx itself); the closure observes the canceled
	// state and decides what to do.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var observedErr error
	err := newTestBlocks().WithSpinner(
		ctx,
		SpinnerSpec{Message: "running"},
		func(_ SpinnerHandle) error {
			observedErr = ctx.Err()
			return nil
		},
	)
	require.NoError(t, err)
	require.ErrorIs(t, observedErr, context.Canceled)
}

func TestWithSpinner_ClosureCompletesInBoundedTime(t *testing.T) {
	// Belt-and-braces: WithSpinner must return promptly once the
	// closure returns, not hang waiting for some unrelated signal.
	done := make(chan struct{})
	go func() {
		_ = newTestBlocks().WithSpinner(
			context.Background(),
			SpinnerSpec{Message: "quick"},
			func(_ SpinnerHandle) error { return nil },
		)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WithSpinner did not return within 2s of closure completing")
	}
}
