package uxHelpers

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
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
		processList[i].spinner = spinners[i]
		wg.Add(1)
		go func(process *Process) {
			defer wg.Done()
			err := process.F(ctx, process)
			if err != nil {
				if process.ErrorMessageMessage == "" {
					send(uxBlock.Finnish())
				} else {
					send(uxBlock.FinnishWithLine(styles.ErrorLine(process.ErrorMessageMessage).String()))
				}
				once.Do(func() {
					returnErr = err
				})
				return
			}
			if process.SuccessMessage == "" {
				send(uxBlock.Finnish())
			} else {
				send(uxBlock.FinnishWithLine(styles.SuccessLine(process.SuccessMessage).String()))
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

func CheckZeropsProcess(
	processId uuid.ProcessId,
	restApiClient *zeropsRestApiClient.Handler,
	options ...gn.Option[checkZeropsProcessSetup],
) ProcessFunc {
	setup := gn.ApplyOptions(options...)
	return func(ctx context.Context, process *Process) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				getProcessResponse, err := restApiClient.GetProcess(ctx, path.ProcessId{Id: processId})
				if err != nil {
					return err
				}

				processOutput, err := getProcessResponse.Output()
				if err != nil {
					return err
				}

				if callback, exists := setup.processOutputCallback.Get(); exists {
					if err := callback(ctx, process, processOutput); err != nil {
						return err
					}
				}

				processStatus := processOutput.Status

				switch processStatus {
				case enum.ProcessStatusEnumPending:
					continue
				case enum.ProcessStatusEnumRunning:
					continue
				case enum.ProcessStatusEnumFinished:
					return nil
				case enum.ProcessStatusEnumRollbacking:
					fallthrough
				case enum.ProcessStatusEnumCanceling:
					fallthrough
				case enum.ProcessStatusEnumFailed:
					fallthrough
				case enum.ProcessStatusEnumCanceled:
					fallthrough
				default:
					return errors.Errorf(i18n.T(i18n.ProcessInvalidState), processId)
				}
			}
		}
	}
}
