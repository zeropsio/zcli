package processChecker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
)

func CheckMultiple(ctx context.Context, process []string, apiGrpcClient zBusinessZeropsApiProtocol.ZBusinessZeropsApiProtocolClient, wg *sync.WaitGroup, sp *spinner.Spinner) {
	processId, name := getProcessData(process)
	isRunning := false

	s := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
	s.HideCursor = true
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			getProcessResponse, err := apiGrpcClient.GetProcess(ctx, &zBusinessZeropsApiProtocol.GetProcessRequest{
				Id: processId,
			})
			if err := proto.BusinessError(getProcessResponse, err); err != nil {
				fmt.Printf("\n%s failed: %v", name, err)
				return
			}

			processStatus := getProcessResponse.GetOutput().GetStatus()

			if processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_RUNNING {
				if !isRunning {
					// stop initial progress indicator that waits for first running process
					if sp.Active() {
						sp.Stop()
						fmt.Println(i18n.ReadyToImportServices)
					}
					clearLine()
					fmt.Printf("%s%s %s \n", constants.Starting, name, i18n.ProcessStart)
					isRunning = true
					// start current service progress indicator
					s.Start()
				}
			}

			if processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_FINISHED {
				s.Stop()
				clearLine()
				fmt.Printf("%s%s %s\n", constants.Success, name, i18n.ProcessEnd)
				return
			}

			if !(processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_RUNNING ||
				processStatus == zBusinessZeropsApiProtocol.ProcessStatus_PROCESS_STATUS_PENDING) {
				s.Stop()
				clearLine()
				fmt.Printf("! %s %s %s\n", name, i18n.ProcessInvalidStateProcess, processId)
				return
			}
			time.Sleep(time.Second)
		}
	}
}

// clear process indicator leftover when interrupted by another process
func clearLine() {
	// 6 spaces equals max number of | chars in spinner to work for Windows
	_, _ = fmt.Fprint(os.Stdout, "\r      \r")
}

func getProcessData(process []string) (string, string) {
	processId := process[0]
	action := strings.Split(process[2], ".")[1] // stack.actionName => actionName
	name := process[1] + " " + action           // e.g. app create
	return processId, name
}
