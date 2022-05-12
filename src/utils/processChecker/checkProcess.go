package processChecker

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func CheckProcess(ctx context.Context, processId string, apiGrpcClient business.ZeropsApiProtocolClient) error {
	sp := spinner.New(spinner.CharSets[33], 100*time.Millisecond)
	sp.Start()
	for {
		select {
		case <-ctx.Done():
			sp.Stop()
			return nil
		default:
			getProcessResponse, err := apiGrpcClient.GetProcess(ctx, &business.GetProcessRequest{
				Id: processId,
			})
			if err := proto.BusinessError(getProcessResponse, err); err != nil {
				return err
			}

			processStatus := getProcessResponse.GetOutput().GetStatus()

			if processStatus == business.ProcessStatus_PROCESS_STATUS_FINISHED {
				sp.Stop()
				return nil
			}

			if !(processStatus == business.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == business.ProcessStatus_PROCESS_STATUS_PENDING) {
				sp.Stop()
				return fmt.Errorf(i18n.ProcessInvalidState, getProcessResponse.GetOutput().GetId())
			}
			time.Sleep(time.Second)
		}
	}
}
