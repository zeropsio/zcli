package importProjectService

import (
	// "bytes"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/zerops-io/zcli/src/i18n"
)

func (h *Handler) ImportService(ctx context.Context, config RunConfig) error {

	fmt.Println("start checking yaml")
	// vyhledat uvedený soubor, pokud nebyl nalezen, zobrazit chybu
	importYamlContent, err := getImportYamlContent(config)
	if err != nil {
		return err
	}

	if len(importYamlContent) == 0 {
		return errors.New(i18n.ImportYamlCorrupted)
	}

	fmt.Printf("config content is %v \n\n END \n\n", string(importYamlContent))

	// vyhledat projekt pomocí gRPC API /project/search aplikovat filtr podle názvu projektu na přesnou shodu.
	projectId, err := h.getProjectId(ctx, config)
	if err != nil {
		fmt.Println("error ", err)
		return err
	}

	fmt.Println(projectId)
	// provést request na gRPC ekvivalent endpointu POST /service-stack/import zBusiness API
	// projectId = project.id, yaml = {obsah souboru načteného výše}

	type ImportService struct {
		Id   string
		Yaml []byte
	}

	jsonData, err := json.Marshal(ImportService{Id: projectId, Yaml: importYamlContent})
	if err != nil {
		return err
	}

	res, err := h.httpClient.Post("/api/rest/public/service-stack/import", jsonData)
	if err != nil {
		return err
	}

	fmt.Println(res)

	// type stack struct {
	//                              Error interface{}
	//                              Id string
	//                              Name string
	//                              Processes [
	//                                {
	//                                  ActionName string,
	//                                  AppVersion interface{},
	//                                  CanceledByUser interface{},
	//                                  ClientId string,
	//                                  Created string,
	//                                  CreatedBySystem boolean,
	//                                  CreatedByUser interface{},
	//                                  Finished string,
	//                                  Id string,
	//                                  LastUpdate string,
	//                                  Project interface{},
	//                                  ProjectId string,
	//                                  Sequence int,
	//                                  ServiceStackId string,
	//                                  ServiceStacks [
	//                                    {
	//                                      Created string,
	//                                      DriverId string,
	//                                      Id string,
	//                                      LastUpdate string,
	//                                      Name string,
	//                                      Ports interface,
	//                                      ProjectId string,
	//                                      ServiceStackTypeId string,
	//                                      ServiceStackTypeInfo interface{},
	//                                      ServiceStackTypeVersionId string
	//                                    }
	//                                  ],
	//                                  Started string,
	//                                  Status string
	//                                }
	//                              ]
	//                            }
	// type importServiceResp struct {
	//   ProjectId string,
	//   ProjectName string,
	//   ServiceStacks stack[]
	// }
	// pokud API vrátilo chybu zobrazit ji; Pozor, některé validační chyby se mohou týkat různých stacků a jsou vráceny v poli serviceStacks[].exception
	// Tyto výjimky zobrazit všechny společně s názvem stacku. V případě, že je u výjimky exception.meta zobrazit u každé výjimky navíc obsah tohoto parametru.

	if res.StatusCode != http.StatusOK {
		// var jsonData []importServiceResp

		// err = json.Unmarshal([]byte(jsonDataFromHttp), &jsonData)
		// if err != nil {
		// 	panic(err)
		// }
		// // TODO find serviceStacks[].exception
		// fmt.Println(serviceStacks)
		return errors.New(i18n.ImportServiceFailed)
	}

	// pokud API vrátilo success, zobrazit informaci o zahájení importu stacků. V textu uvést počet stacků, které se budou vytvářet na základě počtu záznamů v poli serviceStacks[], které vrátilo API
	/////////////////////

	//provádět opakované dotazy na seznam procesů pomocí gRPC API /process/search
	// aplikovat filtr na seznam ID procesů vrácených v serviceStacks[].processes[].id
	// deployProcessId := deployResponse.GetOutput().GetId()

	// err = h.checkProcess(ctx, deployProcessId)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(i18n.BuildDeploySuccess)

	return nil
}
