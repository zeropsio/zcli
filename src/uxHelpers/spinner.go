package uxHelpers

import (
	"context"
	"io"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// Handle is the live spinner context handed to ProcessFunc / ProcessCallback.
// It exposes a writer for streaming log output into the spinner's log view
// (or to stdout when not attached to a terminal).
type Handle interface {
	LogView(opts ...logView.Option) io.Writer
}

type ProcessFunc func(ctx context.Context, h Handle) error
type ProcessCallback func(ctx context.Context, h Handle, apiProcess output.Process) error

type Process struct {
	F              ProcessFunc
	RunningMessage string
	ErrorMessage   string
	SuccessMessage string
}

type spinnerHandle struct {
	spinner *uxBlock.Spinner
}

func (h *spinnerHandle) LogView(opts ...logView.Option) io.Writer {
	if !terminal.IsTerminal() {
		return os.Stdout
	}
	return h.spinner.LogView(opts...)
}

// RunProcessWithSpinner runs a single process with a spinner and returns its
// error. Use RunProcessesWithSpinner for the parallel form.
func RunProcessWithSpinner(ctx context.Context, uxBlocks *uxBlock.Blocks, process Process) error {
	return RunProcessesWithSpinner(ctx, uxBlocks, []Process{process})
}

// RunProcessesWithSpinner runs every process in parallel, each with its own
// spinner. If any process returns an error, the shared context is canceled so
// siblings observe the cancellation; the first non-nil error is returned.
func RunProcessesWithSpinner(ctx context.Context, uxBlocks *uxBlock.Blocks, processList []Process) error {
	if len(processList) == 0 {
		return nil
	}
	spinners := make([]*uxBlock.Spinner, 0, len(processList))
	for _, process := range processList {
		spinners = append(spinners,
			uxBlock.NewSpinner(styles.NewLine(styles.InfoText(process.RunningMessage)).String()),
		)
	}
	stopFunc, send := uxBlocks.RunSpinners(ctx, spinners)
	defer stopFunc()

	g, gctx := errgroup.WithContext(ctx)
	for i := range processList {
		spinner := spinners[i]
		process := processList[i]
		g.Go(func() error {
			err := process.F(gctx, &spinnerHandle{spinner: spinner})
			if err != nil {
				if process.ErrorMessage == "" {
					send(spinner.Finish())
				} else {
					send(spinner.FinishWithLine(styles.ErrorLine(process.ErrorMessage).String()))
				}
				return err
			}
			if process.SuccessMessage == "" {
				send(spinner.Finish())
			} else {
				send(spinner.FinishWithLine(styles.SuccessLine(process.SuccessMessage).String()))
			}
			return nil
		})
	}
	return g.Wait()
}

func CheckZeropsProcessWithProcessOutputCallback(callback ProcessCallback) gn.Option[checkZeropsProcessSetup] {
	return func(c *checkZeropsProcessSetup) {
		c.processOutputCallback = optional.New(callback)
	}
}

type checkZeropsProcessSetup struct {
	processOutputCallback optional.Null[ProcessCallback]
}

func CheckZeropsProcess(
	processId uuid.ProcessId,
	restApiClient *zeropsRestApiClient.Handler,
	uxBlocks *uxBlock.Blocks,
	options ...gn.Option[checkZeropsProcessSetup],
) ProcessFunc {
	setup := gn.ApplyOptions(options...)
	return func(ctx context.Context, h Handle) error {
		var cb func(context.Context, output.Process) error
		if raw, exists := setup.processOutputCallback.Get(); exists {
			cb = func(c context.Context, p output.Process) error {
				return raw(c, h, p)
			}
		}
		return PollProcess(ctx, restApiClient, uxBlocks, processId, cb)
	}
}
