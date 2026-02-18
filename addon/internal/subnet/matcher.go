package subnet

import (
	"net"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
)

type network struct {
	cidr   string
	prefix int
	net    *net.IPNet
}

type Matcher struct {
	networks []network
}

func New() *Matcher {
	return &Matcher{}
}

func (m *Matcher) WithAddresses(addresses []routeros.IPAddress) *Matcher {
	nets := make([]network, 0, len(addresses))
	for _, addr := range addresses {
		cidr := strings.TrimSpace(addr.Address)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		prefix, _ := ipNet.Mask.Size()
		nets = append(nets, network{cidr: cidr, prefix: prefix, net: ipNet})
	}
	return &Matcher{networks: nets}
}

func (m *Matcher) Match(ipStr string) string {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return ""
	}
	bestPrefix := -1
	best := ""
	for _, network := range m.networks {
		if network.net.Contains(ip) && network.prefix > bestPrefix {
			bestPrefix = network.prefix
			best = network.cidr
		}
	}
	return best
}
