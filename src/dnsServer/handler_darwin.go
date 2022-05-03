package dnsServer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/zerops-io/zcli/src/nettools"

	"github.com/miekg/dns"
)

var (
	lo0IP = net.ParseIP("127.0.0.99")
)

type Handler struct {
	lock              sync.RWMutex
	address           net.IP
	forwardAddress    []net.IP
	vpnForwardAddress *net.UDPAddr
	vpnForward        bool
	vpnNetwork        net.IPNet
	ptrPrefix         string
	dnsClient         dns.Client
}

func New() *Handler {
	h := &Handler{
		forwardAddress: []net.IP{
			net.ParseIP("8.8.8.8"),
		},
		dnsClient: dns.Client{
			ReadTimeout:  time.Second * 10,
			WriteTimeout: time.Second * 10,
		},
	}
	return h
}

// listen and serve
func (h *Handler) Run(ctx context.Context) error {
	_, vpnInterfaceFound, err := nettools.GetInterfaceNameByIp(lo0IP)
	if err != nil {
		return err
	}
	if !vpnInterfaceFound {
		vpnInterfaceName := "lo0"
		c := exec.Command("ifconfig", vpnInterfaceName, "alias", lo0IP.String(), "up")
		if err := c.Run(); err != nil {
			return fmt.Errorf("error create %s interface alias on %s: %s", vpnInterfaceName, lo0IP.String(), err.Error())
		}
	}

	listenAddr := &net.UDPAddr{
		IP:   lo0IP,
		Port: 53,
	}
	time.Sleep(time.Second * 5)
	serverUdp := &dns.Server{
		Addr:    listenAddr.String(),
		Net:     "udp",
		Handler: h,
	}
	doneUdp := make(chan struct{})
	go func() {
		defer close(doneUdp)
		defer fmt.Println("stop: ", listenAddr.String())
		fmt.Println("Listen: ", listenAddr.String())
		if err := serverUdp.ListenAndServe(); err != nil {
			fmt.Println(err.Error())
		}
		vpnInterfaceName, vpnInterfaceFound, err := nettools.GetInterfaceNameByIp(lo0IP)
		if err != nil {
			fmt.Println(err.Error())
		}
		if vpnInterfaceFound {
			c := exec.Command("ifconfig", vpnInterfaceName, "delete", lo0IP.String())
			if err := c.Run(); err != nil {
				fmt.Println(err.Error())
			}
		}
	}()
	<-ctx.Done()
	if err := serverUdp.Shutdown(); err != nil {
		fmt.Println(err.Error())
	}
	<-doneUdp
	return nil
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	defer w.Close()

	response, err := h.parseQuery(ctx, r)
	if err != nil {
		fmt.Println("parseQuery", err.Error())
		return
	}

	if err := w.WriteMsg(response); err != nil {
		fmt.Println("writeMsg", err.Error())
		return
	}
}

func (h *Handler) parseQuery(ctx context.Context, in *dns.Msg) (out *dns.Msg, err error) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(in.Question) == 0 {
		return

	}
	if len(in.Question) > 1 {
		out.Rcode = dns.RcodeServerFailure
		return
	}

	q := in.Question[0]

	fmt.Println(q.Name, dns.TypeToString[q.Qtype])
	if h.vpnForward && strings.HasSuffix(q.Name, ".zerops.") || strings.HasSuffix(q.Name, h.ptrPrefix) {
		if q.Qtype == dns.TypeA {
			m := new(dns.Msg)
			m.SetRcode(in, dns.RcodeServerFailure)
			return m, nil
		}
		source := q.Name
		response, err := h.serveVpnForward(ctx, in)
		if err != nil {
			return nil, err
		}
		for _, res := range response.Answer {
			fmt.Println("RESPONSE ", source, "[", dns.TypeToString[q.Qtype], "] -> ", res.String(), res.Header().Rrtype)
		}
		return response, err
	}
	return h.serveForward(ctx, in)
}
func (h *Handler) serveForward(ctx context.Context, m *dns.Msg) (*dns.Msg, error) {
	for _, server := range h.forwardAddress {
		address := &net.UDPAddr{
			IP:   server,
			Port: 53,
		}
		in, _, err := h.dnsClient.ExchangeContext(ctx, m, address.String())
		if err != nil {
			fmt.Println("forward error", m.Question[0].Name, err, address.String())
			continue
		}
		return in, nil
	}
	return nil, errors.New("forward error")
}

func (h *Handler) serveVpnForward(ctx context.Context, m *dns.Msg) (*dns.Msg, error) {
	in, _, err := h.dnsClient.ExchangeContext(ctx, m, h.vpnForwardAddress.String())
	if err != nil {
		fmt.Println("vpn forward", err, h.vpnForwardAddress.String())
		return nil, err
	}
	return in, err
}

func (h *Handler) StopForward() {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.vpnForward = false
}

func (h *Handler) SetAddresses(serverAddress net.IP, userResolverIp []net.IP, vpnResolverIp net.IP, vpnNetwork net.IPNet) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.address = serverAddress
	h.forwardAddress = userResolverIp
	h.vpnForward = true
	h.vpnNetwork = vpnNetwork
	if ptr, err := dns.ReverseAddr(vpnNetwork.IP.String()); err == nil {
		ones, bits := vpnNetwork.Mask.Size()
		h.ptrPrefix = "." + ptr[(bits-ones)/4*2:]

	}
	h.vpnForwardAddress = &net.UDPAddr{
		IP:   vpnResolverIp,
		Port: 53,
	}
	fmt.Println("vpnNetwork: ", h.vpnNetwork.String())
	fmt.Println("address: ", h.address)
	fmt.Println("forwardAddress: ", h.forwardAddress)
	fmt.Println("vpnForwardAddress: ", h.vpnForwardAddress)
	fmt.Println("ptrPrefix: ", h.ptrPrefix)
}
