package nettools

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"

	"github.com/zeropsio/zcli/src/i18n"
)

const ProtocolIPv6ICMP = 58

func Ping(ctx context.Context, address string) error {
	deadline, has := ctx.Deadline()
	if !has {
		return errors.New(i18n.InternalServerError)
	}

	connection, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		return err
	}
	defer connection.Close()

	dst, err := net.ResolveIPAddr("ip6", address)
	if err != nil {
		return err
	}

	message := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte(""),
		},
	}

	buffer, err := message.Marshal(nil)
	if err != nil {
		return err
	}

	n, err := connection.WriteTo(buffer, dst)
	if err != nil {
		return err
	} else if n != len(buffer) {
		return fmt.Errorf("only %d bytes written, want %d bytes written", n, len(buffer))
	}

	err = connection.SetReadDeadline(deadline)
	if err != nil {
		return err
	}

	reply := make([]byte, 1500)
	n, peer, err := connection.ReadFrom(reply)
	if err != nil {
		return err
	}

	response, err := icmp.ParseMessage(ProtocolIPv6ICMP, reply[:n])
	if err != nil {
		return err
	}

	local, err := net.ResolveIPAddr("ip6", "::")
	if err != nil {
		return err
	}

	if !checkAddresses(peer, dst, local) {
		return errors.New(i18n.VpnStatusCheckInvalidAddress)
	}

	if response.Type != ipv6.ICMPTypeEchoReply {
		return fmt.Errorf("got %+v from %v; want echo reply", response, peer)
	}

	return nil
}

func checkAddresses(read net.Addr, write ...net.Addr) bool {
	readIp := getIp(read)

	for _, w := range write {
		writeIp := getIp(w)

		if writeIp.String() == readIp.String() {
			return true
		}
	}
	return false
}

func getIp(addr net.Addr) net.IP {
	switch addr := addr.(type) {
	case *net.IPAddr:
		return addr.IP
	default:
		return net.IPv6zero
	}
}
