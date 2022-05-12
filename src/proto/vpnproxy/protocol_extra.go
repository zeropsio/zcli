package vpnproxy

import (
	"time"

	"net"
)

func FromProtoIP(m *IP) net.IP {
	return m.GetAddress()
}

func ToProtoIP(ip net.IP) *IP {
	return &IP{
		Address: ip,
	}
}

func ToProtoIpNet(ipNet *net.IPNet) *IPRange {
	return &IPRange{
		Ip:   ipNet.IP,
		Mask: ipNet.Mask,
	}
}

func ToProtoTimestamp(t time.Time) *Timestamp {
	if t.IsZero() {
		return &Timestamp{}
	}
	return &Timestamp{IsSet: true, Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}

func FromProtoTimestamp(t *Timestamp) time.Time {
	if !t.GetIsSet() {
		return time.Time{}
	}
	return time.Unix(t.GetSeconds(), int64(t.GetNanos()))
}

func FromProtoIPRange(m *IPRange) net.IPNet {
	return net.IPNet{
		IP:   m.GetIp(),
		Mask: m.GetMask(),
	}
}

func IsInternal(resp interface{}) bool {
	errorGetter, ok := resp.(interface{ GetError() *Error })
	if !ok {
		return true
	}
	return errorGetter.GetError().GetCode() == ErrorCode_INTERNAL_SERVER_ERROR
}

func (x *Error) GetCodeInt() int {
	return int(x.GetCode())
}

func (x *Error) GetMeta() []byte {
	return nil
}
