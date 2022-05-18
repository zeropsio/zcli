package importProjectService

import (
	"fmt"

	"github.com/zerops-io/zcli/src/proto/business"
)

// return number of services and process data [](process Id, service name, action name)
func parseServiceData(servicesData []*business.ProjectImportServiceStack) (int, [][]string) {
	var (
		serviceErrors []*business.Error // TODO do we need to collect them or just to print??
		serviceNames  []string
		processData   [][]string
	)

	for _, service := range servicesData {
		serviceErr := service.GetError().GetValue()
		if serviceErr != nil {
			meta := ""
			if len(serviceErr.GetMeta()) > 0 {
				meta = fmt.Sprintf("\n%v", serviceErr.GetMeta())
			}
			fmt.Printf("service %s returned error %s%s", service.GetName(), serviceErr.GetMessage(), meta)
			serviceErrors = append(serviceErrors, serviceErr)
		}

		serviceNames = append(serviceNames, service.GetName())
		processes := service.GetProcesses()

		for _, process := range processes {
			processData = append(processData, []string{process.GetId(), service.GetName(), process.GetActionName()})
		}
	}
	return len(serviceNames), processData
}
