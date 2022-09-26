package vpn

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"sort"
	"strconv"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dns"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/nettools"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/vpnproxy"
	"github.com/zerops-io/zcli/src/proto/zBusinessZeropsApiProtocol"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	preferredPortMin uint32,
	preferredPortMax uint32,
) (err error) {
	defer func() {
		if err != nil {
			h.logger.Error(err)
			cleanErr := h.stopVpn(ctx)
			if cleanErr != nil {
				h.logger.Error(cleanErr)
			}
		}
	}()

	if err := h.stopVpn(ctx); err != nil {
		return err
	}

	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return err
	}

	apiClientFactory := zBusinessZeropsApiProtocol.New(zBusinessZeropsApiProtocol.Config{CaCertificateUrl: caCertificateUrl})
	apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(ctx, grpcApiAddress, token)
	if err != nil {
		return err
	}
	defer closeFunc()

	h.logger.Debug("vpn request start")

	apiVpnRequestResponse, err := apiGrpcClient.PostVpnRequest(ctx, &zBusinessZeropsApiProtocol.PostVpnRequestRequest{
		Id:              projectId,
		ClientPublicKey: privateKey.PublicKey().String(),
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
	h.logger.Debug("preferredPortMin: ", preferredPortMin)
	h.logger.Debug("preferredPortMax: ", preferredPortMax)

	startVpnResponse, err := vpnGrpcClient.StartVpn(ctx, &vpnproxy.StartVpnRequest{
		AccessToken:      accessToken,
		PreferredPortMin: preferredPortMin,
		PreferredPortMax: preferredPortMax,
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

	if err := h.setVpn(ctx, vpnAddress, privateKey, mtu, startVpnResponse); err != nil {
		return err
	}

	if err := h.updateVpnInterfaceName(privateKey); err != nil {
		return err
	}

	dnsManagement, err := dns.DetectDns()
	if err != nil {
		h.logger.Error(err)
		return err
	}

	vpnNetwork := net.IPNet{
		IP:   startVpnResponse.GetVpn().VpnIpRange.GetIp(),
		Mask: startVpnResponse.GetVpn().VpnIpRange.GetMask(),
	}

	dnsIp := vpnproxy.FromProtoIP(startVpnResponse.GetVpn().GetDnsIp())

	data, err := h.storage.Update(func(data daemonStorage.Data) daemonStorage.Data {
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
		data.DnsManagement = dnsManagement
		data.CaCertificateUrl = caCertificateUrl
		data.Expiry = zBusinessZeropsApiProtocol.FromProtoTimestamp(expiry)
		return data
	})
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.DaemonUnableToSaveConfiguration)
	}

	err = dns.SetDns(data, h.dnsServer)
	if err != nil {
		h.logger.Error(err)
		return err
	}

	h.logger.Debug("dnsIp: " + data.DnsIp.String())
	h.logger.Debug("clientIp: " + data.ClientIp.String())
	h.logger.Debug("dnsManagementType: " + data.DnsManagement)
	h.logger.Debug("serverIp: " + data.ServerIp.String())
	h.logger.Debug("vpnNetwork: " + data.VpnNetwork.String())
	h.logger.Debug("interface: " + data.InterfaceName)

	h.logger.Debug("try vpn")
	if !h.isVpnTunnelAlive(ctx, serverIp) {
		if err := h.stopVpn(ctx); err != nil {
			h.logger.Error(err)
			return err
		}
		return errors.New(i18n.VpnStartTunnelIsNotAlive)
	}
	return nil
}

func (h *Handler) updateVpnInterfaceName(privateKey wgtypes.Key) error {
	wgClient, err := wgctrl.New()
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wgClient.Close()

	wgDevices, err := wgClient.Devices()
	if err != nil {
		h.logger.Error(err)
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	for _, device := range wgDevices {
		device := device
		if device.PrivateKey.String() == privateKey.String() {
			h.storage.Update(func(data daemonStorage.Data) daemonStorage.Data {
				data.InterfaceName = device.Name
				h.logger.Info("set device", data.InterfaceName)
				return data
			})
			return nil
		}
	}
	return errors.New(i18n.VpnStatusDnsInterfaceNotFound)
}
