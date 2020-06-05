package helpers

import "net"

func IpToString(ip net.IP) string {
	if ip.To16() != nil {
		return "[" + ip.String() + "]"
	}

	return ip.String()
}
