package processChecker

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
)

func CheckProcess(ctx context.Context, processId string, apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient) error {
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
	sp.Start()
	defer sp.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			getProcessResponse, err := apiGrpcClient.GetProcess(ctx, &zBusinessZeropsApiProtocol.GetProcessRequest{
				Id: processId,
			})
			if err := proto.BusinessError(getProcessResponse, err); err != nil {
				return err
			}

			processStatus := getProcessResponse.GetOutput().GetStatus()

			if processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_FINISHED {
				return nil
			}

			if !(processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_PENDING) {
				return fmt.Errorf(i18n.ProcessInvalidState, getProcessResponse.GetOutput().GetId())
			}
			time.Sleep(time.Second)
		}
	}
}
