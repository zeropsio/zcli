package vpn

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/utils/certReader"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
)

func (h *Handler) startVpn(
	ctx context.Context,
	grpcApiAddress string,
	grpcVpnAddress string,
	token string,
	projectId string,
	mtu uint32,
) (err error) {
	defer func() {
		if err != nil {
			h.logger.Error(err)
		}
	}()

	err = h.cleanVpn()
	if err != nil {
		return err
	}

	publicKey, privateKey, err := h.generateKeys()
	if err != nil {
		return err
	}

	certReader, err := certReader.New(
		certReader.Config{
			Token: token,
		},
	)
	if err != nil {
		return err
	}

	apiGrpcClient, closeFunc, err := h.grpcApiClientFactory.CreateClient(ctx, grpcApiAddress, token)
	if err != nil {
		return err
	}
	defer closeFunc()

	h.logger.Debug("vpn request start")

	apiVpnRequestResponse, err := apiGrpcClient.PostVpnRequest(ctx, &zeropsApiProtocol.PostVpnRequestRequest{
		Id:              projectId,
		ClientPublicKey: publicKey,
	})
	if err := utils.HandleGrpcApiError(apiVpnRequestResponse, err); err != nil {
		return err
	}
	h.logger.Debug("vpn request end")

	expiry := zeropsApiProtocol.FromProtoTimestamp(apiVpnRequestResponse.GetOutput().GetExpiry())
	signature := apiVpnRequestResponse.GetOutput().GetSignature()

	h.logger.Debug("get vpn addresses start")

	ipRecords, err := net.LookupIP(grpcVpnAddress)
	if err != nil {
		return err
	}

	h.logger.Debug("get vpn addresses end")

	vpnAddress := ""
	for _, ip := range ipRecords {
		vpnAddress = utils.IpToString(ip)
		break
	}

	vpnGrpcClient, closeFunc, err := h.startVpnClient(ctx, vpnAddress)
	if err != nil {
		return err
	}
	defer closeFunc()

	h.logger.Debug("call start vpn - start")

	startVpnResponse, err := vpnGrpcClient.StartVpn(ctx, &zeropsVpnProtocol.StartVpnRequest{
		InstanceId:      projectId,
		UserId:          certReader.UserId,
		ClientPublicKey: publicKey,
		Signature:       signature,
		Expiry:          zeropsVpnProtocol.ToProtoTimestamp(expiry),
	})
	if err := utils.HandleVpnApiError(startVpnResponse, err); err != nil {
		return err
	}

	h.logger.Debug("call start vpn - end")

	clientIp := zeropsVpnProtocol.FromProtoIP(startVpnResponse.GetVpn().GetAssignedClientIp())
	vpnRange := zeropsVpnProtocol.FromProtoIPRange(startVpnResponse.GetVpn().GetVpnIpRange())
	serverIp := zeropsVpnProtocol.FromProtoIP(startVpnResponse.GetVpn().GetServerIp())

	h.logger.Debug("assigned client address: " + clientIp.String())
	h.logger.Debug("assigned vpn server: " + vpnAddress + ":" + strconv.Itoa(int(startVpnResponse.GetVpn().GetPort())))
	h.logger.Debug("server public key: " + startVpnResponse.GetVpn().GetServerPublicKey())
	h.logger.Debug("serverIp address: " + serverIp.String())
	h.logger.Debug("vpnRange: " + vpnRange.String())
	h.logger.Debug("mtu: " + strconv.Itoa(int(mtu)))

	err = h.setVpn(vpnAddress, privateKey, mtu, startVpnResponse)
	if err != nil {
		return err
	}

	dnsManagement, err := h.detectDns()
	if err != nil {
		return err
	}

	vpnNetwork := net.IPNet{
		IP:   startVpnResponse.GetVpn().VpnIpRange.GetIp(),
		Mask: startVpnResponse.GetVpn().VpnIpRange.GetMask(),
	}

	dnsIp := zeropsVpnProtocol.FromProtoIP(startVpnResponse.GetVpn().GetDnsIp())
	h.logger.Debug("dnsIp: " + dnsIp.String())
	h.logger.Debug("clientIp: " + clientIp.String())
	h.logger.Debug("dnsManagementType: " + dnsManagement)
	h.logger.Debug("serverIp: " + serverIp.String())
	h.logger.Debug("vpnNetwork: " + vpnNetwork.String())

	err = h.setDns(dnsIp, clientIp, vpnNetwork, dnsManagement)
	if err != nil {
		return err
	}

	h.logger.Debug("try vpn")
	if !h.isVpnAlive(serverIp.String()) {
		return errors.New("vpn is not connected")
	}

	data := h.storage.Data()
	data.ServerIp = serverIp.String()
	data.VpnNetwork = vpnNetwork.String()
	data.ProjectId = projectId
	data.Mtu = mtu
	data.DnsIp = dnsIp
	data.ClientIp = clientIp
	data.GrpcApiAddress = grpcApiAddress
	data.GrpcVpnAddress = grpcVpnAddress
	data.Token = token

	err = h.storage.Save(data)
	if err != nil {
		return err
	}

	return nil
}
