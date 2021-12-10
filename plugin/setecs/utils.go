package setecs

import (
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strings"

	"github.com/c-robinson/iplib"
	"github.com/miekg/dns"
)

// 解析 IP 到网络对象
func parseIpNet(d string) (inet iplib.Net, err error) {
	if !strings.Contains(d, "/") {
		d = d + "/32"
	}
	_net := iplib.Net4FromStr(d)
	if _net.IP() == nil {
		_net6 := iplib.Net6FromStr(d)
		if _net6.IP() == nil {
			return nil, fmt.Errorf("error ip  %s", d)
		}
		return _net6, nil
	}
	return _net, nil
}

func newEDNS0Subnet(ip net.IP, mask uint8, v6 bool) *dns.EDNS0_SUBNET {
	edns0Subnet := new(dns.EDNS0_SUBNET)
	// edns family: https://www.iana.org/assignments/address-family-numbers/address-family-numbers.xhtml
	// ipv4 = 1
	// ipv6 = 2
	if !v6 { // ipv4
		edns0Subnet.Family = 1
	} else { // ipv6
		edns0Subnet.Family = 2
	}

	edns0Subnet.SourceNetmask = mask
	edns0Subnet.Code = dns.EDNS0SUBNET
	edns0Subnet.Address = ip

	// SCOPE PREFIX-LENGTH, an unsigned octet representing the leftmost
	// number of significant bits of ADDRESS that the response covers.
	// In queries, it MUST be set to 0.
	// https://tools.ietf.org/html/rfc7871
	edns0Subnet.SourceScope = 0
	return edns0Subnet
}

func setECS(m *dns.Msg, ecs *dns.EDNS0_SUBNET) *dns.Msg {
	opt := m.IsEdns0()
	if opt == nil { // no opt, we need a new opt
		o := new(dns.OPT)
		o.SetUDPSize(dns.MinMsgSize)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
		o.Option = []dns.EDNS0{ecs}
		m.Extra = append(m.Extra, o)
		return m
	}

	// if m has a opt, search ecs section
	for o := range opt.Option {
		if opt.Option[o].Option() == dns.EDNS0SUBNET { // overwrite
			opt.Option[o] = ecs
			return m
		}
	}

	// no ecs section, append it
	opt.Option = append(opt.Option, ecs)
	return m
}

func getMsgECS(m *dns.Msg) (e *dns.EDNS0_SUBNET) {
	opt := m.IsEdns0()
	if opt == nil { // no opt, no ecs
		return nil
	}
	// find ecs in opt
	for o := range opt.Option {
		if opt.Option[o].Option() == dns.EDNS0SUBNET {
			return opt.Option[o].(*dns.EDNS0_SUBNET)
		}
	}
	return nil
}

func removeECS(m *dns.Msg) (removedECS *dns.EDNS0_SUBNET) {
	opt := m.IsEdns0()
	if opt == nil { // no opt, no ecs
		return nil
	}

	for i := range opt.Option {
		if opt.Option[i].Option() == dns.EDNS0SUBNET {
			removedECS = opt.Option[i].(*dns.EDNS0_SUBNET)
			opt.Option = append(opt.Option[:i], opt.Option[i+1:]...)
			return
		}
	}
	return nil
}

func ParseIpNet(d string) (inet iplib.Net, err error) {
	if !strings.Contains(d, "/") {
		d = d + "/32"
	}
	_net := iplib.Net4FromStr(d)
	if _net.IP() == nil {
		_net6 := iplib.Net6FromStr(d)
		if _net6.IP() == nil {
			return nil, fmt.Errorf("error ip  %s", d)
		}
		return _net6, nil
	}
	return _net, nil
}

func FileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func StringHash(str string) uint64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}
	return h.Sum64()
}
