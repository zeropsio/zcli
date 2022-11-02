package dns

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zeropsio/zcli/src/daemonStorage"
	"github.com/zeropsio/zcli/src/dnsServer"
)

var serviceOrderNameRegExp = regexp.MustCompile("^\\(([0-9*]+)\\) (.*)$")
var serviceOrderPortRegExp = regexp.MustCompile("^\\(Hardware Port: ([^,]+), Device: ([^)]+)\\)$")

type service struct {
	Index           int
	Name            string
	InterfaceName   string
	InterfaceActive bool
	Disabled        bool
}

func getServiceOrder() (result []service, _ error) {
	output, err := exec.Command("networksetup", "-listnetworkserviceorder").Output()
	if err != nil {
		return nil, err
	}
	dnsScan := bufio.NewScanner(bytes.NewReader(output))
	for dnsScan.Scan() {
		line := dnsScan.Text()
		if match := serviceOrderNameRegExp.FindStringSubmatch(line); len(match) > 0 {
			index := 99
			disabled := true
			if match[1] != "*" {
				index, err = strconv.Atoi(match[1])
				if err != nil {
					continue
				}
				disabled = false
			}
			ser := service{
				Index:    index,
				Name:     match[2],
				Disabled: disabled,
			}
			if dnsScan.Scan() {
				line := dnsScan.Text()
				if match := serviceOrderPortRegExp.FindStringSubmatch(line); len(match) > 0 {
					ser.InterfaceName = match[2]
					in, err := net.InterfaceByName(match[2])
					if err != nil {
						continue
					}
					ser.InterfaceActive = (in.Flags & net.FlagUp) > 0
					result = append(result, ser)
				}
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Index > result[j].Index
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
		if s.InterfaceActive {
			ser = s
		}
	}

	if !ser.InterfaceActive {
		return nil, errors.New("unable to find active network service")
	}

	var ipv6Enabled bool
	if addZerops {
		infoOutput, err := exec.Command("networksetup", "-getinfo", ser.Name).Output()
		if err != nil {
			return nil, err
		}
		{
			infoScan := bufio.NewScanner(bytes.NewReader(infoOutput))
			for infoScan.Scan() {
				infoText := infoScan.Text()

				if strings.HasPrefix(infoText, "IPv6:") {
					if strings.TrimSpace(strings.TrimPrefix(infoText, "IPv6:")) != "Off" {
						ipv6Enabled = true
					}
				}
			}
			dataUpdate = func(data daemonStorage.Data) daemonStorage.Data {
				data.IPv6Enabled = ipv6Enabled
				return data
			}
		}
	} else {
		ipv6Enabled = data.IPv6Enabled
	}
	if !ipv6Enabled {
		if addZerops {
			size, _ := data.VpnNetwork.Mask.Size()
			args := []string{
				"-setv6manual",
				ser.Name,
				data.ClientIp.String(),
				strconv.Itoa(size),
				data.ServerIp.String(),
			}
			if output, err := exec.Command("networksetup", args...).CombinedOutput(); err != nil {
				return nil, fmt.Errorf("unable to set ipv6 routing %v: %s\n\n %s", args, err.Error(), string(output))
			}
		} else {
			args := []string{
				"-setv6off",
				ser.Name,
			}
			if output, err := exec.Command("networksetup", args...).CombinedOutput(); err != nil {
				return nil, fmt.Errorf("unable to unset ipv6 routing %v: %s\n\n %s", args, err.Error(), string(output))
			}

		}
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
			if output, err := exec.Command("networksetup", args...).CombinedOutput(); err != nil {
				return nil, fmt.Errorf("unable to set dnsservers %v: %s\n\n %s", args, err.Error(), string(output))
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
					args = append(args, "Empty")
				} else {
					args = append(args, searchDomains...)
				}
			}
			if output, err := exec.Command("networksetup", args...).CombinedOutput(); err != nil {
				return dataUpdate, fmt.Errorf("unable to set searchdomain %v: %s\n\n %s", args, err.Error(), string(output))
			}
		}
	}

	if !addZerops {
		dns.StopForward()
	}
	return dataUpdate, nil
}
