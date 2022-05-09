package nettools

import (
	"net"
	"time"
)

func PickIP(port string, ips ...net.IP) net.IP {
	timeout := time.Second * 5
	for _, ip := range ips {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), port), timeout)
		if err != nil {
			continue
		}
		conn.Close()
		return ip
	}
	return nil
}
