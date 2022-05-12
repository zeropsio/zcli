package importProjectService

import (
	// "bytes"
	"context"
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	// "github.com/zerops-io/zcli/src/proto"
	// "github.com/zerops-io/zcli/src/proto/business"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	importYamlContent, err := getImportYamlContent(config)
	if err != nil {
		return err
	}

	if len(importYamlContent) == 0 {
		return errors.New(i18n.ImportYamlCorrupted)
	}

	fmt.Printf("config content is %v \n\n END \n\n", string(importYamlContent))

	// res, err := h.apiGrpcClient.PostServiceImport(ctx, &business.PostServiceImportRequest{
	// 	ProjectId: projectId,
	// 	Yaml:     string(importYamlContent),
	// })
	// if err := proto.BusinessError(res, err); err != nil {
	// 	return err
	// }

	// fmt.Println("RESPONSE", res)

	// processId := deployResponse.GetOutput().GetId()

	// err = processChecker.CheckProcess(ctx, deployProcessId, h.apiGrpcClient)
	//	if err != nil {
	//		return err
	//	}

	fmt.Println(constants.Success + i18n.ProjectImportSuccess)

	return nil
}
