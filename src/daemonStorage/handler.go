package daemonStorage

import (
	"net"
	"time"

	"github.com/zerops-io/zcli/src/utils/storage"
)

type Handler struct {
	*storage.Handler[Data]
}

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
	Expiry           time.Time
	PublicKey        string
	PrivateKey       string
}
