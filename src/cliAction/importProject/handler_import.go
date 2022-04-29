package importProject

import (
	// "bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) ImportProjct(ctx context.Context, config RunConfig) error {

	// TODO add grpc getter to rettrieve it
	clientId := "TIHDoACjQAaFB732qDvaww"
	fmt.Println("start checking")

	configContent, err := getConfigContent(config)
	if err != nil {
		return err
	}
	
	// TODO use some of the constants
	if len(configContent) == 0 {
		return errors.New("import yaml corrupted")
	}

	fmt.Printf("config content is %v \n\n END \n\n", string(configContent))

	res, err := h.apiGrpcClient.PostProjectImport(ctx, &zeropsApiProtocol.PostProjectImportRequest{
		ClientId: clientId,
		Yaml: string(configContent),
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

func getConfigContent(config RunConfig) ([]byte, error) {
	workingDir, err := filepath.Abs(config.WorkingDir)
	if err != nil {
		return nil, err
	}

	fmt.Println("working dir is", workingDir)

	if config.ImportYamlPath == nil {
		return nil, errors.New("no path to yaml")
	}

	fmt.Println("yaml path", *config.ImportYamlPath)
	
	importYamlPath := path.Join(workingDir, *config.ImportYamlPath)
	fmt.Println("PATH ",importYamlPath)

	importYamlStat, err := os.Stat(importYamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			if config.ImportYamlPath != nil {
				return nil, errors.New(i18n.ImportYamlNotFound)
			}
		}
		return nil, nil
	}

	fmt.Printf("%s: %s\n", i18n.ImportYamlFound, importYamlPath)

	if importYamlStat.Size() == 0 {
		return nil, errors.New(i18n.ImportYamlEmpty)
	}
	// TODO ask if the size is ok for this yaml (might be larger than zerops.yaml)
	if importYamlStat.Size() > 10*1024 {
		return nil, errors.New(i18n.ImportYamlTooLarge)
	}

	yamlContent, err := os.ReadFile(importYamlPath)
	if err != nil {
		return nil, err
	}

	return yamlContent, nil
}
