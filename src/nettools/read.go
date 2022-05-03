package nettools

import "net"

func GetInterfaceNameByIp(interfaceIp net.IP) (string, bool, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", false, err
	}
	for _, in := range interfaces {
		addresses, err := in.Addrs()
		if err != nil {
			return "", false, err
		}
		for _, address := range addresses {
			if ip, isIp := address.(*net.IPNet); isIp {
				if ip.IP.Equal(interfaceIp) {
					return in.Name, true, nil
				}
			}
		}
	}
	return "", false, nil
}
