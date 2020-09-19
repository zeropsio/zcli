package vpn

import (
	"fmt"
	"net"

	"github.com/zerops-io/zcli/src/scutil"
)

const (
	ZeropsServiceName = "ZeropsVPN"
)

type ZeropsDynamicStorage struct {
	Active           bool
	VpnInterfaceName string
	VpnNetwork       string
	DomainName       string
	ServerAddresses  []net.IP
	ClientIp         net.IP
	DnsIp            net.IP
	PrimaryService   string
}

type ZeropsServiceDns struct {
	scutil.ServiceDns
	ZeropsVPN bool
}

type ZeropsServiceIPv6 struct {
	ZeropsVPN bool
}

func (z *ZeropsDynamicStorage) Read() error {
	return scutil.UnmarshalKey(fmt.Sprintf("State:/Network/Service/%s", ZeropsServiceName), z)
}

func (z *ZeropsDynamicStorage) Apply() error {
	fmt.Println("z.Active", z.Active)
	fmt.Println("z.PrimaryService: ", z.PrimaryService)
	fmt.Println("z.ServerAddresses: ", z.ServerAddresses)
	fmt.Println("z.DomainName: ", z.DomainName)
	fmt.Println("z.ClientIp: ", z.ClientIp)
	fmt.Println("z.DnsIp: ", z.DnsIp)

	newGlobal := scutil.NetworkGlobalIPv4{}
	scutil.UnmarshalKey("State:/Network/Global/IPv4", &newGlobal)

	// if primary servvice changed or ZeropsVPN is not active - restore old settings
	if newGlobal.PrimaryService != z.PrimaryService || !z.Active && z.PrimaryService != "" {
		// restore old primary service
		fmt.Printf("restore old primary service %s\n", z.PrimaryService)
		oldServiceDns := ZeropsServiceDns{}
		scutil.UnmarshalKey(fmt.Sprintf("State:/Network/Service/%s/DNS", z.PrimaryService), &oldServiceDns)

		oldServiceIPv6 := ZeropsServiceIPv6{}
		scutil.UnmarshalKey(fmt.Sprintf("State:/Network/Service/%s/IPv6", z.PrimaryService), &oldServiceIPv6)

		fmt.Println("oldServiceDns.ZeropsVPN: ", oldServiceDns.ZeropsVPN)

		if oldServiceIPv6.ZeropsVPN {
			scutil.RemoveKey(fmt.Sprintf("State:/Network/Service/%s/IPv6", z.PrimaryService))
		}

		if oldServiceDns.ZeropsVPN {
			if len(z.ServerAddresses) == 0 {
				z.ServerAddresses = []net.IP{net.ParseIP("8.8.8.8")}
			}
			scutil.ChangeKey(fmt.Sprintf("State:/Network/Service/%s/DNS", z.PrimaryService),
				scutil.KeyValue{
					Key:   "DomainName",
					Value: z.DomainName,
				},
				scutil.KeyValue{
					Key:   "ServerAddresses",
					Value: scutil.IPsToArrayValue(z.ServerAddresses...),
				},
				scutil.KeyValue{
					Key:    ZeropsServiceName,
					Delete: true,
				},
			)
		}
	}
	z.PrimaryService = newGlobal.PrimaryService
	newServiceDns := ZeropsServiceDns{}
	scutil.UnmarshalKey(fmt.Sprintf("State:/Network/Service/%s/DNS", z.PrimaryService), &newServiceDns)
	if !newServiceDns.ZeropsVPN {
		z.DomainName = newServiceDns.DomainName
		z.ServerAddresses = newServiceDns.ServerAddresses
	}
	z.Store()

	if !z.Active {
		return nil
	}

	scutil.ChangeKey(fmt.Sprintf("State:/Network/Service/%s/DNS", z.PrimaryService),
		scutil.KeyValue{
			Key:   "DomainName",
			Value: "zerops",
		},
		scutil.KeyValue{
			Key:   "ServerAddresses",
			Value: "* 127.0.0.99",
		},
		scutil.KeyValue{
			Key:   ZeropsServiceName,
			Value: "true",
		},
	)
	if !scutil.KeyExists(fmt.Sprintf("State:/Network/Service/%s/IPv6", z.PrimaryService)) {
		fmt.Println("IPv6 connectivity does not exists")
		scutil.MoveKey(
			fmt.Sprintf("State:/Network/Interface/%s/IPv6", z.VpnInterfaceName),
			fmt.Sprintf("State:/Network/Service/%s/IPv6", z.PrimaryService),
			scutil.KeyValue{
				Key:   "Router",
				Value: z.DnsIp.String(),
			},
			scutil.KeyValue{
				Key:   "InterfaceName",
				Value: z.VpnInterfaceName,
			},
			scutil.KeyValue{
				Key:   ZeropsServiceName,
				Value: "true",
			},
		)
	} else {
		fmt.Println("IPv6 connectivity exists")
	}
	return nil
}

func (z ZeropsDynamicStorage) activeValue() string {
	if z.Active {
		return "true"
	}
	return "false"
}

func (z ZeropsDynamicStorage) Store() error {
	return scutil.ChangeKey(fmt.Sprintf("State:/Network/Service/%s", ZeropsServiceName),
		scutil.KeyValue{
			Key:   "DomainName",
			Value: z.DomainName,
		},
		scutil.KeyValue{
			Key:   "Active",
			Value: z.activeValue(),
		},
		scutil.KeyValue{
			Key:   "ServerAddresses",
			Value: scutil.IPsToArrayValue(z.ServerAddresses...),
		},
		scutil.KeyValue{
			Key:   "VpnInterfaceName",
			Value: z.VpnInterfaceName,
		},
		scutil.KeyValue{
			Key:   "ClientIp",
			Value: z.ClientIp.String(),
		},
		scutil.KeyValue{
			Key:   "DnsIp",
			Value: z.DnsIp.String(),
		},
		scutil.KeyValue{
			Key:   "PrimaryService",
			Value: z.PrimaryService,
		},
	)
}
