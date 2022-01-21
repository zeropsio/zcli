package utils

import "net"

func IpToString(ip net.IP) string {
	if ip.To4() == nil {
		return "[" + ip.String() + "]"
	}

	return ip.String()
}
