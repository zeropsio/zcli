package vpn

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/zerops-io/zcli/src/utils/interfaces"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/grpcApiClientFactory"
	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
	"google.golang.org/grpc/status"
)

func (h *Handler) startVpn(
	ctx context.Context,
	grpcApiAddress string,
	grpcVpnAddress string,
	token string,
	projectId string,
	userId string,
	mtu uint32,
	caCertificateUrl string,
) (err error) {
	defer func() {
		if err != nil {
			h.logger.Error(err)
			err = h.cleanVpn()
			if err != nil {
				h.logger.Error(err)
			}
		}
	}()

	data := h.storage.Data()
	if data.VpnStarted {
		return nil
	}

	publicKey, privateKey, err := h.generateKeys()
	if err != nil {
		return err
	}

	apiClientFactory := grpcApiClientFactory.New(grpcApiClientFactory.Config{CaCertificateUrl: caCertificateUrl})
	apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(ctx, grpcApiAddress, token)
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
		if errStatus, ok := status.FromError(err); ok {
			return errors.New(errStatus.Err().Error())
		} else {
			return err
		}
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

	sort.Slice(ipRecords, func(i, j int) bool { return rand.Int()%2 == 0 })

	vpnAddress := ""
	for _, ip := range ipRecords {
		ipString := utils.IpToString(ip)
		conn, err := net.DialTimeout("tcp", ipString+vpnApiGrpcPort, 5*time.Second)
		if err != nil {
			h.logger.Debug("check vpn addresses: " + ipString + " failed " + err.Error())
			continue
		}
		conn.Close()
		h.logger.Debug("check vpn addresses: " + ipString + " success")
		vpnAddress = ipString
		break
	}

	if vpnAddress == "" {
		return errors.New(i18n.VpnStartVpnNotReachable)
	}

	vpnGrpcClient, closeFunc, err := h.startVpnClient(ctx, vpnAddress)
	if err != nil {
		return err
	}
	defer closeFunc()

	h.logger.Debug("call start vpn - start")

	startVpnResponse, err := vpnGrpcClient.StartVpn(ctx, &zeropsVpnProtocol.StartVpnRequest{
		InstanceId:      projectId,
		UserId:          userId,
		ClientPublicKey: publicKey,
		Signature:       signature,
		Expiry:          zeropsVpnProtocol.ToProtoTimestamp(expiry),
	})
	if err := utils.HandleVpnApiError(startVpnResponse, err); err != nil {
		if errStatus, ok := status.FromError(err); ok {
			return errors.New(errStatus.Err().Error())
		} else {
			return err
		}
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

	dnsManagement, err := dns.DetectDns()
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

	err = dns.SetDns(h.dnsServer, dnsIp, clientIp, vpnNetwork, dnsManagement)
	if err != nil {
		return err
	}

	ifName, _, err := interfaces.GetInterfaceNameByIp(clientIp)
	if err != nil {
		return err
	}

	h.logger.Debug("try vpn")
	if !h.isVpnTunnelAlive(serverIp) {
		dns.CleanDns(h.dnsServer, dnsIp, ifName, dnsManagement)
		return errors.New(i18n.VpnStartTunnelIsNotAlive)
	}

	data.ServerIp = serverIp
	data.VpnNetwork = vpnNetwork
	data.ProjectId = projectId
	data.UserId = userId
	data.Mtu = mtu
	data.DnsIp = dnsIp
	data.ClientIp = clientIp
	data.GrpcApiAddress = grpcApiAddress
	data.GrpcVpnAddress = grpcVpnAddress
	data.Token = token
	data.DnsManagement = string(dnsManagement)
	data.CaCertificateUrl = caCertificateUrl
	data.VpnStarted = true
	data.InterfaceName = ifName

	err = h.storage.Save(data)
	if err != nil {
		return err
	}

	return nil
}
