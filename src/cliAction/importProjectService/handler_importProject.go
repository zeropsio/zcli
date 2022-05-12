package importProjectService

import (
	// "bytes"
	"context"
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	// todo replace with more relevant message /==> checking yaml/
	fmt.Println("start checking")

	importYamlContent, err := getImportYamlContent(config)
	if err != nil {
		return err
	}

	if len(importYamlContent) == 0 {
		return errors.New(i18n.ImportYamlCorrupted)
	}

	clientId, err := h.getClientId(ctx, config)
	if err != nil {
		return err
	}

	res, err := h.apiGrpcClient.PostProjectImport(ctx, &business.PostProjectImportRequest{
		ClientId: clientId,
		Yaml:     string(importYamlContent),
	})
	if err := proto.BusinessError(res, err); err != nil {
		return err
	}

	fmt.Println("RESPONSE", res)
	// ProjectId     string                       `protobuf:"bytes,1,opt,name=projectId,proto3" json:"projectId,omitempty"`
	//	ProjectName   string                       `protobuf:"bytes,2,opt,name=projectName,proto3" json:"projectName,omitempty"`
	//	ServiceStacks
	serviceStacksData := res.GetOutput().GetServiceStacks()
	fmt.Println(serviceStacksData)
	//ProjectImportServiceStack:
	//Id        string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	//	Name      string     `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	//	Error     *ErrorNull `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	//	Processes []*Process `protobuf:"bytes,4,rep,name=processes,proto3" json:"processes,omitempty"`

	// processId := deployResponse.GetOutput().GetId()

	// err = processChecker.CheckProcess(ctx, deployProcessId, h.apiGrpcClient)
	//	if err != nil {
	//		return err
	//	}

	fmt.Println(constants.Success + i18n.ProjectImportSuccess)

	return nil
}
