package uxHelpers

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// ProcessMessages are the three lines surfaced by WatchProcess as a
// single zerops process moves from pending to finished. Each is
// optional; an empty SuccessMessage leaves the spinner with no
// trailing line; an empty ErrorMessage falls back to err.Error().
type ProcessMessages struct {
	Running string
	Success string
	Error   string
}

// WatchProcess polls a zerops process to completion behind a single
// spinner. Replaces the older []Process / ProcessCheckWithSpinner
// pattern for the common single-process case:
//
//	uxHelpers.WatchProcess(ctx, blocks, client, processId, uxHelpers.ProcessMessages{
//	    Running: i18n.T(i18n.ServiceStopping),
//	    Error:   i18n.T(i18n.ServiceStopFailed),
//	    Success: i18n.T(i18n.ServiceStopped),
//	})
func WatchProcess(
	ctx context.Context,
	blocks *uxBlock.Blocks,
	client *zeropsRestApiClient.Handler,
	processId uuid.ProcessId,
	messages ProcessMessages,
) error {
	return blocks.WithSpinner(ctx, uxBlock.SpinnerSpec{
		Message:        messages.Running,
		SuccessMessage: messages.Success,
		ErrorMessage:   messages.Error,
	}, func(_ uxBlock.SpinnerHandle) error {
		return pollProcess(ctx, client, blocks, processId, nil)
	})
}

// pollProcess is the polling loop shared by WatchProcess and the
// legacy CheckZeropsProcess. callback, if non-nil, is invoked on every
// successful poll with the latest process snapshot — callers that
// stream phase-specific logs use this to attach loggers as phases
// transition. Transient API errors are tolerated for up to one minute
// before being surfaced.
func pollProcess(
	ctx context.Context,
	client *zeropsRestApiClient.Handler,
	blocks *uxBlock.Blocks,
	processId uuid.ProcessId,
	callback func(ctx context.Context, p output.Process) error,
) error {
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
		blocks.Printf(styles.WarningLine, "failed to get process detail: %v", err)
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			resp, err := client.GetProcess(ctx, path.ProcessId{Id: processId})
			if err != nil {
				if err := handleTransientErr(err); err != nil {
					return err
				}
				continue
			}

			processOutput, err := resp.Output()
			if err != nil {
				if err := handleTransientErr(err); err != nil {
					return err
				}
				continue
			}
			firstErrTime = time.Time{}

			if callback != nil {
				if err := callback(ctx, processOutput); err != nil {
					return err
				}
			}

			switch processOutput.Status {
			case enum.ProcessStatusEnumPending, enum.ProcessStatusEnumRunning:
				continue
			case enum.ProcessStatusEnumFinished:
				return nil
			default:
				return errors.Errorf(i18n.T(i18n.ProcessInvalidState), processId)
			}
		}
	}
}
