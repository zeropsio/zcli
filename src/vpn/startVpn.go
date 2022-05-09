package vpn

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"sort"
	"strconv"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/nettools"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/proto/vpnproxy"
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
			cleanErr := h.cleanVpn()
			if cleanErr != nil {
				h.logger.Error(cleanErr)
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

	apiClientFactory := business.New(business.Config{CaCertificateUrl: caCertificateUrl})
	apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(ctx, grpcApiAddress, token)
	if err != nil {
		return err
	}
	defer closeFunc()

	h.logger.Debug("vpn request start")

	apiVpnRequestResponse, err := apiGrpcClient.PostVpnRequest(ctx, &business.PostVpnRequestRequest{
		Id:              projectId,
		ClientPublicKey: publicKey,
	})
	if err := proto.BusinessError(apiVpnRequestResponse, err); err != nil {
		return err
	}
	h.logger.Debug("vpn request end")

	accessToken := apiVpnRequestResponse.GetOutput().GetAccessToken()
	expiry := apiVpnRequestResponse.GetOutput().GetExpiry()

	h.logger.Debug("get vpn addresses start")

	ipRecords, err := net.LookupIP(grpcVpnAddress)
	if err != nil {
		return err
	}

	h.logger.Debug("get vpn addresses end")

	sort.Slice(ipRecords, func(i, j int) bool { return rand.Int()%2 == 0 })

	vpnAddress := nettools.PickIP(constants.VpnApiGrpcPort, ipRecords...)
	if vpnAddress == nil {
		return errors.New(i18n.VpnStartVpnNotReachable)
	}

	targetVpnAddress := net.JoinHostPort(vpnAddress.String(), constants.VpnApiGrpcPort)

	vpnGrpcClient, closeFunc, err := vpnproxy.CreateClient(ctx, targetVpnAddress)
	if err != nil {
		return err
	}
	defer closeFunc()

	h.logger.Debug("call start vpn - start")

	startVpnResponse, err := vpnGrpcClient.StartVpn(ctx, &vpnproxy.StartVpnRequest{
		AccessToken: accessToken,
	})

	if err := proto.VpnError(startVpnResponse, err); err != nil {
		return err
	}

	h.logger.Debug("call start vpn - end")

	clientIp := vpnproxy.FromProtoIP(startVpnResponse.GetVpn().GetAssignedClientIp())
	vpnRange := vpnproxy.FromProtoIPRange(startVpnResponse.GetVpn().GetVpnIpRange())
	serverIp := vpnproxy.FromProtoIP(startVpnResponse.GetVpn().GetServerIp())

	h.logger.Debug("assigned client address: " + clientIp.String())
	h.logger.Debug("assigned vpn server: " + vpnAddress.String() + ":" + strconv.Itoa(int(startVpnResponse.GetVpn().GetPort())))
	h.logger.Debug("server public key: " + startVpnResponse.GetVpn().GetServerPublicKey())
	h.logger.Debug("serverIp address: " + serverIp.String())
	h.logger.Debug("vpnRange: " + vpnRange.String())
	h.logger.Debug("mtu: " + strconv.Itoa(int(mtu)))

	vpnPortHostAddress := net.JoinHostPort(vpnAddress.String(), strconv.Itoa(int(startVpnResponse.GetVpn().GetPort())))
	err = h.setVpn(vpnPortHostAddress, privateKey, mtu, startVpnResponse)
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

	dnsIp := vpnproxy.FromProtoIP(startVpnResponse.GetVpn().GetDnsIp())
	h.logger.Debug("dnsIp: " + dnsIp.String())
	h.logger.Debug("clientIp: " + clientIp.String())
	h.logger.Debug("dnsManagementType: " + dnsManagement)
	h.logger.Debug("serverIp: " + serverIp.String())
	h.logger.Debug("vpnNetwork: " + vpnNetwork.String())

	err = dns.SetDns(h.dnsServer, dnsIp, clientIp, vpnNetwork, dnsManagement)
	if err != nil {
		return err
	}

	ifName, _, err := nettools.GetInterfaceNameByIp(clientIp)
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
	data.GrpcVpnAddress = targetVpnAddress
	data.Token = token
	data.DnsManagement = string(dnsManagement)
	data.CaCertificateUrl = caCertificateUrl
	data.VpnStarted = true
	data.InterfaceName = ifName
	data.Expiry = business.FromProtoTimestamp(expiry)
	data.PublicKey = publicKey
	data.PrivateKey = privateKey

	err = h.storage.Save(data)
	if err != nil {
		return err
	}

	return nil
}
