package importProjectService

import (
	"fmt"

	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
)

// return number of services and process data [](process Id, service name, action name)
func parseServiceData(servicesData []*zBusinessZeropsApiProtocol.ProjectImportServiceStack) (int, [][]string) {
	var (
		serviceNames = make([]string, 0, len(servicesData))
		processData  [][]string
	)

	for _, service := range servicesData {
		serviceErr := service.GetError().GetValue()
		if serviceErr != nil {
			meta := ""
			if len(serviceErr.GetMeta()) > 0 {
				meta = fmt.Sprintf("\n%s", string(serviceErr.GetMeta()))
			}
			fmt.Printf("service %s returned error %s%s", service.GetName(), serviceErr.GetMessage(), meta)
		}

		serviceNames = append(serviceNames, service.GetName())
		processes := service.GetProcesses()

		for _, process := range processes {
			processData = append(processData, []string{process.GetId(), service.GetName(), process.GetActionName()})
		}
	}
	return len(serviceNames), processData
}
