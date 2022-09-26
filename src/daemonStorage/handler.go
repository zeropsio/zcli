package daemonStorage

import (
	"net"
	"time"

	"github.com/zerops-io/zcli/src/utils/storage"
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
	LocalDnsManagementScutil         LocalDnsManagement = "SCUTIL"
	LocalDnsManagementUnknown        LocalDnsManagement = "UNKNOWN"
	LocalDnsManagementWindows        LocalDnsManagement = "WINDOWS"
)

type Data struct {
	ProjectId        string
	UserId           string
	VpnNetwork       net.IPNet
	GrpcApiAddress   string
	GrpcVpnAddress   string
	CaCertificateUrl string
	Token            string

	InterfaceName string
	ServerIp      net.IP
	DnsIp         net.IP
	ClientIp      net.IP
	Mtu           uint32
	DnsManagement LocalDnsManagement

	Expiry time.Time
}
