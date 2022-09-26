package dns

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
	"strings"

	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/dnsServer"
)

func setDnsByNetworksetup(data daemonStorage.Data, dns *dnsServer.Handler, addZerops bool) error {
	output, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return err
	}
	serviceScan := bufio.NewScanner(bytes.NewReader(output))
	serviceScanIndex := 0
	var serverAddresses []net.IP
	var services []string
	for serviceScan.Scan() {
		serviceScanIndex++
		serviceName := serviceScan.Text()
		if serviceScanIndex == 1 {
			continue
		}
		services = append(services, serviceName)
		{
			dnsOutput, err := exec.Command("networksetup", "-getdnsservers", serviceName).Output()
			if err != nil {
				return err
			}

			dnsScan := bufio.NewScanner(bytes.NewReader(dnsOutput))
			var dnsServers []net.IP
			for dnsScan.Scan() {
				dnsServer := dnsScan.Text()
				dnsServerIp := net.ParseIP(dnsServer)
				if dnsServerIp.String() == dnsServer && dnsServerIp.String() != "127.0.0.99" {
					dnsServers = append(dnsServers, dnsServerIp)
					serverAddresses = append(serverAddresses, dnsServerIp)
				}

			}
			if len(dnsServers) > 0 {
				args := []string{
					"-setdnsservers", serviceName,
				}
				if addZerops {
					args = append(args, "127.0.0.99")
				}
				for _, dnsIp := range dnsServers {
					args = append(args, dnsIp.String())
				}
				_, err := exec.Command("networksetup", args...).Output()
				if err != nil {
					return err
				}
			}
		}
		{
			searchOutput, err := exec.Command("networksetup", "-getsearchdomains", serviceName).Output()
			if err != nil {
				return err
			}

			var searchDomains []string
			searchScan := bufio.NewScanner(bytes.NewReader(searchOutput))
			for searchScan.Scan() {
				searchDomain := searchScan.Text()
				if searchDomain != "zerops" && !strings.Contains(searchDomain, " ") {
					searchDomains = append(searchDomains, searchDomain)
				}

			}
			if len(searchDomains) > 0 {
				args := []string{
					"-setsearchdomains", serviceName,
				}
				if addZerops {
					args = append(args, "zerops")
				}
				args = append(args, searchDomains...)
				_, err := exec.Command("networksetup", args...).Output()
				if err != nil {
					return err
				}
			}
		}
	}
	if addZerops {
		dns.SetAddresses(
			data.ClientIp,
			serverAddresses,
			data.DnsIp,
			data.VpnNetwork,
		)
	} else {
		dns.StopForward()
	}
	return nil
}
