package importProjectService

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/briandowns/spinner"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/processChecker"
)

func (h *Handler) Import(ctx context.Context, config RunConfig) error {

	importYamlContent, err := getImportYamlContent(config)
	if err != nil {
		return err
	}

	var servicesData []*zBusinessZeropsApiProtocol.ProjectImportServiceStack
	isProjectCmd := config.ParentCmd == constants.Project

	if isProjectCmd {
		servicesData, err = h.sendProjectRequest(ctx, config, string(importYamlContent))
	} else {
		servicesData, err = h.sendServiceRequest(ctx, config, string(importYamlContent))
	}
	if err != nil {
		return err
	}

	serviceCount, processData := parseServiceData(servicesData)

	fmt.Println(i18n.ServiceStackCount + strconv.Itoa(serviceCount))
	fmt.Println(i18n.QueuedProcesses + strconv.Itoa(len(processData)))

	if isProjectCmd {
		fmt.Println(i18n.CoreServices)
	}

	var wg sync.WaitGroup
	wg.Add(len(processData))
	sp := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
	sp.Start()

	for _, processItem := range processData {
		go processChecker.CheckMultiple(ctx, processItem, h.apiGrpcClient, &wg, sp)
	}
	wg.Wait()

	if isProjectCmd {
		fmt.Println(constants.Success + i18n.ProjectImported)
	} else {
		fmt.Println(constants.Success + i18n.ServiceImported)
	}

	return nil
}
