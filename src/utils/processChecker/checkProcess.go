package processChecker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func CheckProcess(ctx context.Context, processId string, apiGrpcClient business.ZeropsApiProtocolClient) error {
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
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

func CheckMultiple(ctx context.Context, process []string, apiGrpcClient business.ZeropsApiProtocolClient, wg *sync.WaitGroup, sp *spinner.Spinner) {
	processId := process[0]
	action := strings.Split(process[2], ".")[1] // stack.actionName => actionName
	name := process[1] + " " + action           // e.g. app create
	isRunning := false

	s := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			getProcessResponse, err := apiGrpcClient.GetProcess(ctx, &business.GetProcessRequest{
				Id: processId,
			})
			if err := proto.BusinessError(getProcessResponse, err); err != nil {
				fmt.Printf("\n%s failed: %v", name, err)
				return
			}

			processStatus := getProcessResponse.GetOutput().GetStatus()

			if processStatus == business.ProcessStatus_PROCESS_STATUS_RUNNING {
				// stop create project progress indicator
				if !isRunning {
					if sp.Active() {
						sp.Stop()
					}
					fmt.Printf("\n%s %s ", name, i18n.ProcessStart)
					isRunning = true
					// start first process progress indicator
					s.Start()
				}
			}

			if processStatus == business.ProcessStatus_PROCESS_STATUS_FINISHED {
				s.Stop()
				fmt.Printf("\n%s%s%s\n", constants.Success, name, i18n.ProcessEnd)
				return
			}

			if !(processStatus == business.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == business.ProcessStatus_PROCESS_STATUS_PENDING) {
				s.Stop()
				processErr := fmt.Errorf(i18n.ProcessInvalidStateProcess, getProcessResponse.GetOutput().GetId())
				fmt.Printf("\n%s %s\n", name, processErr)
				return
			}
			time.Sleep(time.Second)
		}
	}
}
