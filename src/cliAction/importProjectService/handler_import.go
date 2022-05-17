package importProjectService

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/utils/processChecker"
)

func (h *Handler) Import(ctx context.Context, config RunConfig, parentCmd string) error {

	importYamlContent, err := getImportYamlContent(config)
	if err != nil {
		return err
	}

	var servicesData []*business.ProjectImportServiceStack

	if parentCmd == constants.Project {
		servicesData, err = sendProjectRequest(ctx, config, h, string(importYamlContent))
	} else {
		servicesData, err = sendServiceRequest(ctx, config, h, string(importYamlContent))
	}
	if err != nil {
		return err
	}

	var (
		serviceErrors []*business.Error // TODO do we need to collect them or just to print??
		serviceNames  []string
		processData   [][]string
		waitGroup     = sync.WaitGroup{}
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
			processData = append(processData, []string{process.GetId(), service.GetName(), process.GetActionName()})
		}
	}

	fmt.Println(i18n.ServiceStackCount + strconv.Itoa(len(serviceNames)))
	fmt.Println(i18n.QueuedProcesses + strconv.Itoa(len(processData)))

	waitGroup.Add(len(processData))
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
	sp.Start()
	for _, processItem := range processData {
		go processChecker.CheckMultiple(ctx, processItem, h.apiGrpcClient, &waitGroup, sp)
	}
	waitGroup.Wait()

	if parentCmd == constants.Project {
		fmt.Println(constants.Success + i18n.ProjectImportSuccess)
	} else {
		fmt.Println(constants.Success + i18n.ServiceImportSuccess)
	}

	return nil
}
