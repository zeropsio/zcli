package processChecker

import (
	"context"
	"fmt"
	"time"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func CheckProcess(ctx context.Context, processId string, apiGrpcClient business.ZeropsApiProtocolClient) error {
	for {
		select {
		case <-ctx.Done():
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
				return nil
			}

			if !(processStatus == business.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == business.ProcessStatus_PROCESS_STATUS_PENDING) {
				return fmt.Errorf(i18n.ProcessInvalidState, getProcessResponse.GetOutput().GetId())
			}
			time.Sleep(time.Second)
		}
	}
}