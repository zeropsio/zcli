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

	serviceCount, processData := parseServiceData(servicesData)

	fmt.Println(i18n.ServiceStackCount + strconv.Itoa(serviceCount))
	fmt.Println(i18n.QueuedProcesses + strconv.Itoa(len(processData)))

	var wg sync.WaitGroup
	wg.Add(len(processData))
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
	sp.Start()
	for _, processItem := range processData {
		go processChecker.CheckMultiple(ctx, processItem, h.apiGrpcClient, &wg, sp)
	}
	wg.Wait()

	if parentCmd == constants.Project {
		fmt.Println(constants.Success + i18n.ProjectImportSuccess)
	} else {
		fmt.Println(constants.Success + i18n.ServiceImportSuccess)
	}

	return nil
}
