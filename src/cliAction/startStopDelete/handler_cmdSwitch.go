package startStopDelete

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
)

type Method func(ctx context.Context, projectId string, serviceId string) (string, error)

type CmdType struct {
	Start   string
	Finish  string
	Execute Method
}

func (h *Handler) getSwitcher() [][]CmdType {
	switcher := make([][]CmdType, 2)
	switcher[constants.Project] = make([]CmdType, 3)
	switcher[constants.Service] = make([]CmdType, 3)

	switcher[constants.Project][constants.Start] = CmdType{
		Start:   i18n.ProjectStart,
		Finish:  i18n.ProjectStarted,
		Execute: h.ProjectStart,
	}
	switcher[constants.Project][constants.Stop] = CmdType{
		Start:   i18n.ProjectStop,
		Finish:  i18n.ProjectStopped,
		Execute: h.ProjectStop,
	}
	switcher[constants.Project][constants.Delete] = CmdType{
		Start:   i18n.ProjectDelete,
		Finish:  i18n.ProjectDeleted,
		Execute: h.ProjectDelete,
	}
	switcher[constants.Service][constants.Start] = CmdType{
		Start:   i18n.ServiceStart,
		Finish:  i18n.ServiceStarted,
		Execute: h.ServiceStart,
	}
	switcher[constants.Service][constants.Stop] = CmdType{
		Start:   i18n.ServiceStop,
		Finish:  i18n.ServiceStopped,
		Execute: h.ServiceStop,
	}
	switcher[constants.Service][constants.Delete] = CmdType{
		Start:   i18n.ServiceDelete,
		Finish:  i18n.ServiceDeleted,
		Execute: h.ServiceDelete,
	}
	return switcher
}

// get start msg, stop msg and grpc method
func (h *Handler) getCmdProps(parentCmd constants.ParentCmd, childCmd constants.ChildCmd) (string, string, Method) {
	selected := h.getSwitcher()[parentCmd][childCmd]
	return selected.Start, selected.Finish, selected.Execute
}
