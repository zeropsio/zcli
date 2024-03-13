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

	var returnErr error

	var once sync.Once

	wg := sync.WaitGroup{}
	for i := range processList {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			process := processList[i]

			err := process.F(ctx)
			if err != nil {
				spinners[i].Finish(styles.ErrorLine(process.ErrorMessageMessage))

				once.Do(func() {
					returnErr = err
				})
				return
			}
			spinners[i].Finish(styles.SuccessLine(process.SuccessMessage))
		}(i)
	}

	wg.Wait()
	stopFunc()

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

				if processStatus == enum.ProcessStatusEnumFinished {
					return nil
				}

				if !(processStatus == enum.ProcessStatusEnumRunning ||
					processStatus == enum.ProcessStatusEnumPending) {
					return errors.Errorf(i18n.T(i18n.ProcessInvalidState), processId)
				}
			}
		}
	}
}
