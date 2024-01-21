package uxHelpers

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func ProcessCheckWithSpinner(
	ctx context.Context,
	uxBlocks *uxBlock.UxBlocks,
	restApiClient *zeropsRestApiClient.Handler,
	processList []Process,
) error {
	spinners := make([]*uxBlock.Spinner, 0, len(processList))
	for _, process := range processList {
		spinners = append(spinners, uxBlock.NewSpinner(process.RunningMessage))
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

			err := checkProcess(ctx, process.Id, restApiClient)
			if err != nil {
				spinners[i].Finish(uxBlock.NewLine(uxBlock.ErrorIcon, uxBlock.ErrorText(process.ErrorMessageMessage)).String())
				stopFunc()

				once.Do(func() {
					returnErr = err
				})
				return
			}
			spinners[i].Finish(uxBlock.NewLine(uxBlock.SuccessIcon, uxBlock.SuccessText(process.SuccessMessage)).String())
		}(i)
	}

	wg.Wait()
	stopFunc()

	return returnErr
}

type Process struct {
	Id                  uuid.ProcessId
	RunningMessage      string
	ErrorMessageMessage string
	SuccessMessage      string
}

func checkProcess(ctx context.Context, processId uuid.ProcessId, restApiClient *zeropsRestApiClient.Handler) error {
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
