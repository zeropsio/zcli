package importProjectService

import (
	// "bytes"
	"context"
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	clientId := config.ClientId
	if len(clientId) == 0 {
		// TODO add grpc getter to retrieve it from gRPC API /client/search
		clientId = "TIHDoACjQAaFB732qDvaww"
	}

	fmt.Println("start checking")

	importYamlContent, err := getImportYamlContent(config)
	if err != nil {
		return err
	}

	if len(importYamlContent) == 0 {
		// TODO use some of the constants
		return errors.New("import yaml corrupted")
	}

	fmt.Printf("config content is %v \n\n END \n\n", string(importYamlContent))

	res, err := h.apiGrpcClient.PostProjectImport(ctx, &zeropsApiProtocol.PostProjectImportRequest{
		ClientId: clientId,
		Yaml:     string(importYamlContent),
	})
	if err := utils.HandleGrpcApiError(res, err); err != nil {
		fmt.Println("ERRR", err)
		return err
	}

	fmt.Println("RESPONSE", res)
	/////////////////////

	// deployProcessId := deployResponse.GetOutput().GetId()

	// err = h.checkProcess(ctx, deployProcessId)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(i18n.BuildDeploySuccess)

	return nil
}
