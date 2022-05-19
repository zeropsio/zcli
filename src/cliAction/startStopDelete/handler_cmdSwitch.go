package startStopDelete

import (
	"context"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
)

type CB func(ctx context.Context, h *Handler, projectId string, serviceId string) (string, error)

type CmdType struct {
	StartMsg  string
	FinishMsg string
	Callback  CB
}

func (h *Handler) getCmdType(parentCmd constants.ParentCmd, childCmd constants.ChildCmd) (string, string, CB) {
	var messages = map[constants.ParentCmd]map[constants.ChildCmd]CmdType{}
	messages[constants.Project][constants.Start] = CmdType{
		StartMsg:  i18n.ProjectStartProcessInit,
		FinishMsg: i18n.ProjectStartSuccess,
		Callback:  ProjectStart,
	}
	messages[constants.Project][constants.Stop] = CmdType{
		StartMsg:  i18n.ProjectStopProcessInit,
		FinishMsg: i18n.ProjectStopSuccess,
		Callback:  ProjectStop,
	}
	messages[constants.Project][constants.Delete] = CmdType{
		StartMsg:  i18n.ProjectDeleteProcessInit,
		FinishMsg: i18n.ProjectDeleteSuccess,
		Callback:  ProjectDelete,
	}
	messages[constants.Service][constants.Start] = CmdType{
		StartMsg:  i18n.ServiceStartProcessInit,
		FinishMsg: i18n.ServiceStartSuccess,
		Callback:  ServiceStart,
	}
	messages[constants.Service][constants.Stop] = CmdType{
		StartMsg:  i18n.ServiceStopProcessInit,
		FinishMsg: i18n.ServiceStopSuccess,
		Callback:  ServiceStop,
	}
	messages[constants.Service][constants.Delete] = CmdType{
		StartMsg:  i18n.ServiceDeleteProcessInit,
		FinishMsg: i18n.ServiceDeleteSuccess,
		Callback:  ServiceDelete,
	}

	selected := messages[parentCmd][childCmd]
	return selected.StartMsg, selected.FinishMsg, selected.Callback
}
