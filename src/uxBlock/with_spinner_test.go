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

func TestSetHeader_StoresAndClears(t *testing.T) {
	// The header is a simple persistent field; tested via the public
	// setter and observed through the unexported field (same package).
	b := newTestBlocks()
	require.Empty(t, b.header)

	b.SetHeader("project=demo  service=api")
	require.Equal(t, "project=demo  service=api", b.header)

	b.SetHeader("")
	require.Empty(t, b.header, "SetHeader(\"\") clears the header")
}

func TestWithSpinner_HeaderSurvivesAcrossCalls(t *testing.T) {
	// Pin the documented contract: SetHeader is sticky, multiple
	// WithSpinner calls in a row all see it. Verified at the Blocks
	// level since the rendered output is non-terminal mode no-op.
	b := newTestBlocks()
	b.SetHeader("scope")

	for range 3 {
		err := b.WithSpinner(context.Background(), SpinnerSpec{Message: "x"},
			func(_ SpinnerHandle) error { return nil })
		require.NoError(t, err)
		require.Equal(t, "scope", b.header, "header must persist across WithSpinner calls")
	}
}

func TestWithSpinners_EmptyJobs(t *testing.T) {
	err := newTestBlocks().WithSpinners(context.Background(), nil)
	require.NoError(t, err)
}

func TestWithSpinners_AllSucceed(t *testing.T) {
	const n = 4
	var called int32
	jobs := make([]SpinnerJob, n)
	for i := range jobs {
		jobs[i] = SpinnerJob{
			Spec: SpinnerSpec{Message: "j", SuccessMessage: "done"},
			Run: func(_ SpinnerHandle) error {
				atomic.AddInt32(&called, 1)
				return nil
			},
		}
	}
	err := newTestBlocks().WithSpinners(context.Background(), jobs)
	require.NoError(t, err)
	require.Equal(t, int32(n), atomic.LoadInt32(&called))
}

func TestWithSpinners_FirstErrorWins(t *testing.T) {
	// Errgroup semantics: first non-nil error returned. We can't
	// deterministically force which goroutine wins the race, but the
	// returned error must be one of the two we raised.
	errA := errors.New("err-a")
	errB := errors.New("err-b")
	err := newTestBlocks().WithSpinners(context.Background(), []SpinnerJob{
		{Spec: SpinnerSpec{Message: "a"}, Run: func(_ SpinnerHandle) error { return errA }},
		{Spec: SpinnerSpec{Message: "b"}, Run: func(_ SpinnerHandle) error { return errB }},
	})
	require.Error(t, err)
	require.Truef(t, errors.Is(err, errA) || errors.Is(err, errB),
		"expected errA or errB, got %v", err)
}

func TestWithSpinners_AllJobsInvokedEvenWhenOneErrors(t *testing.T) {
	// We don't gate sibling jobs on context — the errgroup-derived
	// ctx is not forwarded today (see comment in WithSpinners). So
	// all jobs run to completion regardless of one erroring. Pin
	// this behavior; revisit if we forward gctx.
	const n = 3
	var called int32
	jobs := make([]SpinnerJob, n)
	jobs[0] = SpinnerJob{Spec: SpinnerSpec{Message: "fail"}, Run: func(_ SpinnerHandle) error {
		atomic.AddInt32(&called, 1)
		return errors.New("boom")
	}}
	for i := 1; i < n; i++ {
		jobs[i] = SpinnerJob{Spec: SpinnerSpec{Message: "ok"}, Run: func(_ SpinnerHandle) error {
			atomic.AddInt32(&called, 1)
			time.Sleep(5 * time.Millisecond)
			return nil
		}}
	}
	err := newTestBlocks().WithSpinners(context.Background(), jobs)
	require.Error(t, err)
	require.Equal(t, int32(n), atomic.LoadInt32(&called))
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
