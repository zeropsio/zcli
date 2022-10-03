package dns

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zeropsio/zcli/src/daemonStorage"
	"github.com/zeropsio/zcli/src/dnsServer"
)

var serviceOrderNameRegExp = regexp.MustCompile("^\\(([0-9]+)\\) (.*)$")
var serviceOrderPortRegExp = regexp.MustCompile("^\\(Hardware Port: ([^,]+), Device: ([^)]+)\\)$")

type service struct {
	Name          string
	InterfaceName string
	Active        bool
}

func getServiceOrder() (result []service, _ error) {
	resultMap := make(map[string]int)
	output, err := exec.Command("networksetup", "-listnetworkserviceorder").Output()
	if err != nil {
		return nil, err
	}
	dnsScan := bufio.NewScanner(bytes.NewReader(output))
	for dnsScan.Scan() {
		line := dnsScan.Text()
		if match := serviceOrderNameRegExp.FindStringSubmatch(line); len(match) > 0 {
			index, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}
			resultMap[match[2]] = index
			result = append(result, service{
				Name: match[2],
			})
		}
		if match := serviceOrderPortRegExp.FindStringSubmatch(line); len(match) > 0 {
			for index, ser := range result {
				if ser.Name == match[1] {
					result[index].InterfaceName = match[2]
					in, err := net.InterfaceByName(match[2])
					if err != nil {
						continue
					}
					if !strings.HasPrefix(in.Name, "en") {
						continue
					}
					result[index].Active = (in.Flags & net.FlagUp) > 0
				}
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		iIndex, exists := resultMap[result[i].Name]
		if !exists {
			return false
		}
		jIndex, exists := resultMap[result[j].Name]
		if !exists {
			return false
		}
		return iIndex < jIndex
	})
	return result, nil
}

func setDnsByNetworksetup(data daemonStorage.Data, dns *dnsServer.Handler, addZerops bool) (dataUpdate func(daemonStorage.Data) daemonStorage.Data, _ error) {

	serviceOrder, err := getServiceOrder()
	if err != nil {
		return nil, err
	}
	if len(serviceOrder) == 0 {
		return nil, errors.New("unable to find active network service")
	}

	remoteDnsServer := "127.0.0.99"
	var serverAddresses []net.IP

	var ser service
	for _, s := range serviceOrder {
		if s.Active {
			ser = s
			break
		}
	}

	if !ser.Active {
		return nil, errors.New("unable to find active network service")
	}

	{
		dnsOutput, err := exec.Command("networksetup", "-getdnsservers", ser.Name).Output()
		if err != nil {
			return nil, err
		}
		var dnsServers []net.IP
		{
			dnsScan := bufio.NewScanner(bytes.NewReader(dnsOutput))
			for dnsScan.Scan() {
				dnsServerText := dnsScan.Text()
				dnsServerIp := net.ParseIP(dnsServerText)
				if dnsServerIp.String() == dnsServerText && dnsServerIp.String() != remoteDnsServer {
					dnsServers = append(dnsServers, dnsServerIp)
					serverAddresses = append(serverAddresses, dnsServerIp)
				}
			}
		}

		if len(dnsServers) == 0 {
			if dnsOutput, err := exec.Command("ipconfig", "getoption", ser.InterfaceName, "domain_name_server").Output(); err == nil {
				dnsScan := bufio.NewScanner(bytes.NewReader(dnsOutput))
				for dnsScan.Scan() {
					dnsServerText := dnsScan.Text()
					dnsServerIp := net.ParseIP(dnsServerText)
					if dnsServerIp.String() == dnsServerText && dnsServerIp.String() != remoteDnsServer {
						serverAddresses = append(serverAddresses, dnsServerIp)
						dnsServers = append(dnsServers, dnsServerIp)
					}
				}
			}
			dataUpdate = func(data daemonStorage.Data) daemonStorage.Data {
				data.DhcpEnabled = true
				return data
			}
		} else {
			dataUpdate = func(data daemonStorage.Data) daemonStorage.Data {
				data.DhcpEnabled = false
				return data
			}
		}

		if len(dnsServers) == 0 {
			return nil, errors.New("unknown DNS configuration")
		}

		{
			args := []string{
				"-setdnsservers", ser.Name,
			}
			if addZerops {
				args = append(args, remoteDnsServer)
				dns.SetAddresses(
					data.ClientIp,
					serverAddresses,
					data.DnsIp,
					data.VpnNetwork,
				)
				for _, dnsIp := range dnsServers {
					args = append(args, dnsIp.String())
				}
			} else {
				if data.DhcpEnabled {
					args = append(args, "empty")
				} else {
					for _, dnsIp := range dnsServers {
						args = append(args, dnsIp.String())
					}

				}
			}
			_, err := exec.Command("networksetup", args...).Output()
			if err != nil {
				return nil, err
			}
		}
	}

	{
		var searchDomains []string
		{
			searchOutput, err := exec.Command("networksetup", "-getsearchdomains", ser.Name).Output()
			if err != nil {
				return nil, err
			}
			searchScan := bufio.NewScanner(bytes.NewReader(searchOutput))
			for searchScan.Scan() {
				searchDomain := searchScan.Text()
				if searchDomain != "zerops" && !strings.Contains(searchDomain, " ") {
					searchDomains = append(searchDomains, searchDomain)
				}
			}
		}
		if len(searchDomains) == 0 {
			searchOutput, err := exec.Command("ipconfig", "getoption", ser.InterfaceName, "domain_name").Output()
			if err != nil {
				return nil, err
			}
			searchScan := bufio.NewScanner(bytes.NewReader(searchOutput))
			for searchScan.Scan() {
				searchDomain := searchScan.Text()
				if searchDomain != "zerops" && !strings.Contains(searchDomain, " ") {
					searchDomains = append(searchDomains, searchDomain)
				}
			}
		}
		{
			args := []string{
				"-setsearchdomains", ser.Name,
			}
			if addZerops {
				args = append(args, "zerops")
				args = append(args, searchDomains...)
			} else {
				if data.DhcpEnabled || len(searchDomains) == 0 {
					args = append(args, "empty")
				} else {
					args = append(args, searchDomains...)
				}
			}
			_, err := exec.Command("networksetup", args...).Output()
			if err != nil {
				return dataUpdate, err
			}
		}
	}

	if !addZerops {
		dns.StopForward()
	}
	return dataUpdate, nil
}
