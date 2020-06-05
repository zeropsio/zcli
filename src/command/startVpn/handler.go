package startVpn

import (
	"context"
	"errors"
	"net"

	"github.com/zerops-io/zcli/src/helpers"

	"github.com/zerops-io/zcli/src/service/logger"
	"github.com/zerops-io/zcli/src/service/storage"
	"github.com/zerops-io/zcli/src/service/sudoers"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

const wireguardPort = "51820"
const vpnApiGrpcPort = ":64510"

type Config struct {
	VpnAddress string
	UserId     string
}

type RunConfig struct {
	ProjectId string
}

type Handler struct {
	config        Config
	logger        *logger.Handler
	apiGrpcClient zeropsApiProtocol.ZeropsApiProtocolClient
	sudoers       *sudoers.Handler
	storage       *storage.Handler
}

func New(
	config Config,
	logger *logger.Handler,
	apiGrpcClient zeropsApiProtocol.ZeropsApiProtocolClient,
	sudoers *sudoers.Handler,
	storage *storage.Handler,
) *Handler {
	return &Handler{
		config:        config,
		logger:        logger,
		apiGrpcClient: apiGrpcClient,
		sudoers:       sudoers,
		storage:       storage,
	}
}

func (h *Handler) Run(ctx context.Context, config RunConfig) error {

	if h.storage.Data.ProjectId != "" && config.ProjectId != h.storage.Data.ProjectId {
		if h.isVpnAlive() {
			return errors.New("vpn is started for another project, use stopVpn first")
		}
	}

	err := h.cleanVpn()
	if err != nil {
		return err
	}

	publicKey, privateKey, err := h.generateKeys()
	if err != nil {
		return err
	}

	apiVpnRequestResponse, err := h.apiGrpcClient.PostVpnRequest(ctx, &zeropsApiProtocol.PostVpnRequestRequest{
		Id:              config.ProjectId,
		ClientPublicKey: publicKey,
	})
	if err := helpers.HandleGrpcApiError(apiVpnRequestResponse, err); err != nil {
		return err
	}

	expiry := zeropsApiProtocol.FromProtoTimestamp(apiVpnRequestResponse.GetOutput().GetExpiry())
	signature := apiVpnRequestResponse.GetOutput().GetSignature()

	ipRecords, err := net.LookupIP(h.config.VpnAddress)
	if err != nil {
		return err
	}

	vpnAddress := ""
	for _, ip := range ipRecords {
		vpnAddress = helpers.IpToString(ip)
		break
	}

	vpnGrpcClient, closeFunc, err := h.startVpnClient(ctx, vpnAddress)
	if err != nil {
		return err
	}
	defer closeFunc()

	startVpnResponse, err := vpnGrpcClient.StartVpn(ctx, &zeropsVpnProtocol.StartVpnRequest{
		InstanceId:      config.ProjectId,
		UserId:          h.config.UserId,
		ClientPublicKey: publicKey,
		Signature:       signature,
		Expiry:          zeropsVpnProtocol.ToProtoTimestamp(expiry),
	})
	if err := helpers.HandleVpnApiError(startVpnResponse, err); err != nil {
		return err
	}

	err = h.setVpn(vpnAddress, privateKey, startVpnResponse)
	if err != nil {
		return err
	}

	h.logger.Info("\nclient is connected \n")

	h.storage.Data.ProjectId = config.ProjectId
	err = h.storage.Save()
	if err != nil {
		return err
	}

	return nil
}
