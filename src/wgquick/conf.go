package wgquick

import "net"

type Config struct {
	ClientPrivateKey string
	ClientAddress    net.IP
	DnsServers       []string
	MTU              int

	ServerPublicKey string
	AllowedIPs      net.IPNet
	ServerAddress   string
}
