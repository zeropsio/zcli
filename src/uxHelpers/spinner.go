package uxHelpers

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

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
	uxBlocks uxBlock.UxBlocks,
	processList []Process,
) error {
	spinners := make([]*uxBlock.Spinner, 0, len(processList))
	for _, process := range processList {
		spinners = append(spinners, uxBlock.NewSpinner(styles.NewLine(styles.InfoText(process.RunningMessage))))
	}

	stopFunc := uxBlocks.RunSpinners(ctx, spinners)
	defer stopFunc()

	var returnErr error
	var once sync.Once

	var wg sync.WaitGroup
	for i := range processList {
		wg.Add(1)
		go func(process Process, spinner *uxBlock.Spinner) {
			defer wg.Done()
			err := process.F(ctx)
			if err != nil {
				if process.ErrorMessageMessage == "" {
					spinner.Finish(styles.NewLine())
				} else {
					spinner.Finish(styles.ErrorLine(process.ErrorMessageMessage))
				}
				once.Do(func() {
					returnErr = err
				})
				return
			}
			if process.SuccessMessage == "" {
				spinner.Finish(styles.NewLine())
				return
			}
			spinner.Finish(styles.SuccessLine(process.SuccessMessage))
		}(processList[i], spinners[i])
	}
	wg.Wait()

	return returnErr
}

type Process struct {
	F                   func(ctx context.Context) error
	RunningMessage      string
	ErrorMessageMessage string
	SuccessMessage      string
}

func CheckZeropsProcess(
	processId uuid.ProcessId,
	restApiClient *zeropsRestApiClient.Handler,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
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
