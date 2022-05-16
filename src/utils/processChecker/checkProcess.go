package processChecker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func CheckProcess(ctx context.Context, processId string, apiGrpcClient business.ZeropsApiProtocolClient) error {
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond) // 33, 32, 14
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

func CheckProcesses(ctx context.Context, processId string, name string, apiGrpcClient business.ZeropsApiProtocolClient, wg *sync.WaitGroup) {
	defer wg.Done()
	isRunning := false

	for {
		select {
		case <-ctx.Done():
			return
		default:
			getProcessResponse, err := apiGrpcClient.GetProcess(ctx, &business.GetProcessRequest{
				Id: processId,
			})
			if err := proto.BusinessError(getProcessResponse, err); err != nil {
				fmt.Println("")
				fmt.Println(name + ":")
				fmt.Println(err)
				return
			}

			processStatus := getProcessResponse.GetOutput().GetStatus()

			if processStatus == business.ProcessStatus_PROCESS_STATUS_RUNNING {
				if !isRunning {
					fmt.Println("")
					fmt.Println(name + " process is running")
					isRunning = true
				}
			}

			if processStatus == business.ProcessStatus_PROCESS_STATUS_FINISHED {
				fmt.Println("")
				fmt.Println(constants.Success + name + " process finished")
				return
			}

			if !(processStatus == business.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == business.ProcessStatus_PROCESS_STATUS_PENDING) {
				fmt.Println("")
				fmt.Print(name + ": ")
				processErr := fmt.Errorf(i18n.ProcessInvalidState, getProcessResponse.GetOutput().GetId())
				fmt.Println(processErr)
				return
			}
			time.Sleep(time.Second)
		}
	}
}
