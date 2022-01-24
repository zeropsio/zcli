package wgquick

import "net"

type Config struct {
	ClientPrivateKey string
	ClientAddress    net.IP
	DnsServers       []string

	ServerPublicKey string
	AllowedIPs      net.IPNet
	ServerAddress   string
}
