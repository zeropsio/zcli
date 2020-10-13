package startVpn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/peterh/liner"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	if config.ProjectName == "" {
		return errors.New(i18n.VpnStartProjectNameIsEmpty)
	}

	projectsResponse, err := h.apiGrpcClient.GetProjectsByName(ctx, &zeropsApiProtocol.GetProjectsByNameRequest{
		Name: config.ProjectName,
	})
	if err := utils.HandleGrpcApiError(projectsResponse, err); err != nil {
		return err
	}

	projectsResponse.GetOutput().GetProjects()

	projects := projectsResponse.GetOutput().GetProjects()
	if len(projects) == 0 {
		return errors.New(i18n.VpnStartProjectNotFound)
	}
	if len(projects) > 1 {
		return errors.New(i18n.VpnStartProjectsWithSameName)
	}
	project := projects[0]

	err = h.tryStartVpn(ctx, project, config)
	if err != nil {
		return err
	}

	fmt.Println(i18n.VpnStartSuccess)

	return nil
}

func (h *Handler) tryStartVpn(ctx context.Context, project *zeropsApiProtocol.Project, config RunConfig) error {

	zeropsDaemonClient, closeFn, err := h.zeropsDaemonClientFactory.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	_, err = zeropsDaemonClient.StartVpn(ctx, &zeropsDaemonProtocol.StartVpnRequest{
		ApiAddress: h.config.GrpcApiAddress,
		VpnAddress: h.config.VpnAddress,
		ProjectId:  project.GetId(),
		Token:      config.Token,
		Mtu:        config.Mtu,
	})
	if err != nil {
		if errStatus, ok := status.FromError(err); ok {
			if errStatus.Code() == codes.Unavailable {
				fmt.Println(i18n.VpnStartDaemonIsUnavailable)

				line := liner.NewLiner()
				defer line.Close()
				line.SetCtrlCAborts(true)

				fmt.Println(i18n.VpnStartInstallDaemonPrompt)
				for {
					if answer, err := line.Prompt("y/n "); err == nil {
						if answer == "n" {
							return errors.New(i18n.VpnStartTerminatedByUser)
						} else if answer == "y" {
							err := h.daemonInstaller.Install()
							if err != nil {
								return err
							}
							fmt.Println(i18n.DaemonInstallSuccess)

							// let's wait for daemon start
							time.Sleep(3 * time.Second)
							return h.tryStartVpn(ctx, project, config)
						} else {
							fmt.Println(i18n.VpnStartUserIsUnableToWriteYorN)
							continue
						}
					} else if err == liner.ErrPromptAborted {
						return errors.New(i18n.VpnStartTerminatedByUser)
					} else {
						return err
					}
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
