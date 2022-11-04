package daemonStorage

import (
	"net"
	"time"

	"github.com/zeropsio/zcli/src/utils/storage"
)

type Handler struct {
	*storage.Handler[Data]
}

type LocalDnsManagement string

const (
	LocalDnsManagementSystemdResolve LocalDnsManagement = "SYSTEMD_RESOLVE"
	LocalDnsManagementResolveConf    LocalDnsManagement = "RESOLVCONF"
	LocalDnsManagementFile           LocalDnsManagement = "FILE"
	LocalDnsManagementNetworkSetup   LocalDnsManagement = "NETWORKSETUP"
	LocalDnsManagementUnknown        LocalDnsManagement = "UNKNOWN"
	LocalDnsManagementWindows        LocalDnsManagement = "WINDOWS"
)

type Data struct {
	ProjectId            string
	UserId               string
	VpnNetwork           net.IPNet
	GrpcApiAddress       string
	GrpcVpnAddress       string
	GrpcTargetVpnAddress string
	CaCertificateUrl     string
	Token                string
	PreferredPortMin     uint32
	PreferredPortMax     uint32

	ServerIp      net.IP
	DnsIp         net.IP
	ClientIp      net.IP
	Mtu           uint32
	DnsManagement LocalDnsManagement

	InterfaceName string

	Expiry time.Time
}
