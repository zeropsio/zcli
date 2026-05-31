package uxHelpers

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/optional"
	"github.com/zeropsio/zcli/src/terminal"
	"github.com/zeropsio/zcli/src/uxBlock/models/logView"
	"github.com/zeropsio/zerops-go/dto/output"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func ProcessCheckWithSpinner(
	ctx context.Context,
	uxBlocks *uxBlock.Blocks,
	processList []Process,
) error {
	spinners := make([]*uxBlock.Spinner, 0, len(processList))
	for _, process := range processList {
		spinners = append(spinners,
			uxBlock.NewSpinner(styles.NewLine(styles.InfoText(process.RunningMessage)).String()),
		)
	}
	stopFunc, send := uxBlocks.RunSpinners(ctx, spinners)
	defer stopFunc()

	var returnErr error
	var once sync.Once

	var wg sync.WaitGroup
	for i := range processList {
		spinner := spinners[i]
		processList[i].spinner = spinner
		wg.Add(1)
		go func(process *Process) {
			defer wg.Done()
			err := process.F(ctx, process)
			if err != nil {
				if process.ErrorMessageMessage == "" {
					send(spinner.Finish())
				} else {
					send(spinner.FinishWithLine(styles.ErrorLine(process.ErrorMessageMessage).String()))
				}
				once.Do(func() {
					returnErr = err
				})
				return
			}
			if process.SuccessMessage == "" {
				send(spinner.Finish())
			} else {
				send(spinner.FinishWithLine(styles.SuccessLine(process.SuccessMessage).String()))
			}
		}(&processList[i])
	}
	wg.Wait()

	return returnErr
}

type ProcessFunc func(ctx context.Context, process *Process) error
type ProcessCallback func(ctx context.Context, process *Process, apiProcess output.Process) error

type Process struct {
	F                   ProcessFunc
	RunningMessage      string
	ErrorMessageMessage string
	SuccessMessage      string
	spinner             *uxBlock.Spinner
}

func (p *Process) LogView(opts ...logView.Option) io.Writer {
	if !terminal.IsTerminal() {
		return os.Stdout
	}
	return p.spinner.LogView(opts...)
}

func CheckZeropsProcessWithProcessOutputCallback(callback ProcessCallback) gn.Option[checkZeropsProcessSetup] {
	return func(c *checkZeropsProcessSetup) {
		c.processOutputCallback = optional.New(callback)
	}
}

type checkZeropsProcessSetup struct {
	processOutputCallback optional.Null[ProcessCallback]
}

type processOutcome int

const (
	// processPolling - the process is still in flight, keep waiting.
	processPolling processOutcome = iota
	// processSucceeded - the process finished cleanly.
	processSucceeded
	// processFailed - the process ended in a non-success terminal state (or an unknown one, treated as failure).
	processFailed
)

func classifyProcessStatus(status enum.ProcessStatusEnum) processOutcome {
	switch status {
	case enum.ProcessStatusEnumPending, enum.ProcessStatusEnumRunning:
		return processPolling
	case enum.ProcessStatusEnumFinished:
		return processSucceeded
	default:
		return processFailed
	}
}

func CheckZeropsProcess(
	processId entity.ProcessId,
	restApiClient *zeropsRestApiClient.Handler,
	uxBlocks *uxBlock.Blocks,
	options ...gn.Option[checkZeropsProcessSetup],
) ProcessFunc {
	setup := gn.ApplyOptions(options...)
	return func(ctx context.Context, process *Process) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		var firstErrTime time.Time
		handleTransientErr := func(err error) error {
			if firstErrTime.IsZero() {
				firstErrTime = time.Now()
			}
			if time.Since(firstErrTime) >= time.Minute {
				return err
			}
			uxBlocks.PrintWarningTextf("failed to get process detail: %v", err)
			return nil
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				getProcessResponse, err := restApiClient.GetProcess(ctx, path.ProcessId{Id: uuid.ProcessId(processId)})
				if err != nil {
					if err := handleTransientErr(err); err != nil {
						return err
					}
					continue
				}

				processOutput, err := getProcessResponse.Output()
				if err != nil {
					if err := handleTransientErr(err); err != nil {
						return err
					}
					continue
				}

				firstErrTime = time.Time{}

				if callback, exists := setup.processOutputCallback.Get(); exists {
					if err := callback(ctx, process, processOutput); err != nil {
						return err
					}
				}

				switch classifyProcessStatus(processOutput.Status) {
				case processPolling:
					continue
				case processSucceeded:
					return nil
				case processFailed:
					fallthrough
				default:
					return errors.Errorf(i18n.T(i18n.ProcessInvalidState), processId)
				}
			}
		}
	}
}
