package daemonStorage

import (
	"github.com/zerops-io/zcli/src/utils/storage"
	"net"
)

type Handler = storage.Handler[Data]

type Data struct {
	ProjectId        string
	UserId           string
	ServerIp         net.IP
	VpnNetwork       net.IPNet
	GrpcApiAddress   string
	GrpcVpnAddress   string
	Token            string
	DnsIp            net.IP
	ClientIp         net.IP
	Mtu              uint32
	DnsManagement    string
	CaCertificateUrl string
	VpnStarted       bool
	InterfaceName    string
}
