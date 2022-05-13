package importProjectService

import (
	// "bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
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

	if res.GetError() != nil {
		fmt.Println("response errors: ", res.GetError().GetMessage())
		// todo confirm if only print or return this error
		//return errors.New(res.GetError().GetMessage())
	}

	fmt.Println(constants.Success + i18n.ProjectCreateSuccess)

	servicesData := res.GetOutput().GetServiceStacks()
	// check errors for each, if error, get service name and value and get error meta
	var (
		serviceErrors []*business.Error
		serviceNames  []string
		processData   [][]string
		// todo this is only for development and  will be delete and the above array used
		processIds []string
		waitGroup  = sync.WaitGroup{}
	)

	for _, service := range servicesData {
		serviceErr := service.GetError().GetValue()
		if serviceErr != nil {
			fmt.Println("service " + service.GetName() + " returned error " + serviceErr.GetMessage() + ". \n " + string(serviceErr.GetMeta()))
			serviceErrors = append(serviceErrors, serviceErr)
		}

		serviceNames = append(serviceNames, service.GetName())
		processes := service.GetProcesses()

		for _, process := range processes {
			processData = append(processData, []string{service.GetName(), process.GetId(), process.GetActionName()})
			processIds = append(processIds, process.GetId())
		}
	}

	fmt.Println(i18n.ServiceStackCount + strconv.Itoa(len(serviceNames)))
	fmt.Println(processData)

	waitGroup.Add(len(processIds))
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond) // 33, 32, 14
	sp.Start()
	for _, p := range processData {
		action := strings.Split(p[2], ".")[1]
		go processChecker.CheckProcesses(ctx, p[1], p[0]+" "+action, h.apiGrpcClient, &waitGroup)
	}
	waitGroup.Wait()
	sp.Stop()
	//provádět opakované dotazy na seznam procesů pomocí gRPC API /process/search
	//aplikovat filtr na seznam ID procesů vrácených v serviceStacks[].processes[].id
	//dokud nejsou všechny vrácené procesy ve stavu FINISHED, FAILED nebo CANCELED

	//pokud se proces poprvé změnil do stavu RUNNING zobrazit informaci o spuštění příslušného procesu.
	//Informace bude obsahovat název stacku a název procesu.
	//
	//pokud se proces poprvé změnil do stavu FINISHED, FAILED nebo CANCELED zobrazit informaci o dokončení příslušného procesu.
	//Informace bude obsahovat název stacku, název procesu a stav procesu.
	//
	//
	//pokud jsou všechny vrácené procesy ve stavu FINISHED, FAILED nebo CANCELED zobrazit informaci o dokončení importu stacků a ukončit algoritmus.

	fmt.Println(constants.Success + i18n.ProjectImportSuccess)

	return nil
}
