//go:build darwin

package facts

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// NetworkingCollector gathers network facts on macOS.
type NetworkingCollector struct{}

func (n *NetworkingCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	if hostname, err := os.Hostname(); err == nil {
		result["networking.hostname"] = hostname
		if fqdn, err := net.LookupCNAME(hostname); err == nil && fqdn != "" {
			result["networking.fqdn"] = strings.TrimSuffix(fqdn, ".")
		} else {
			result["networking.fqdn"] = hostname
		}
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return result, nil
	}

	// On macOS we pick the first non-loopback interface with an IPv4 address
	// as the primary. A more robust approach would parse the routing table.
	var primarySet bool
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		prefix := fmt.Sprintf("networking.interfaces.%s", iface.Name)
		result[prefix+".mac"] = iface.HardwareAddr.String()
		result[prefix+".mtu"] = iface.MTU

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.To4() != nil {
				result[prefix+".ip"] = ipNet.IP.String()
				ones, _ := ipNet.Mask.Size()
				result[prefix+".netmask"] = net.IP(ipNet.Mask).String()
				result[prefix+".cidr"] = fmt.Sprintf("%s/%d", ipNet.IP.String(), ones)
				if !primarySet && iface.Flags&net.FlagUp != 0 {
					result["networking.ip"] = ipNet.IP.String()
					result["networking.mac"] = iface.HardwareAddr.String()
					result["networking.netmask"] = net.IP(ipNet.Mask).String()
					primarySet = true
				}
			} else {
				result[prefix+".ip6"] = ipNet.IP.String()
				if !primarySet || result["networking.ip6"] == nil {
					result["networking.ip6"] = ipNet.IP.String()
				}
			}
		}
	}

	if nameservers := parseResolvConf(); len(nameservers) > 0 {
		result["networking.nameservers"] = nameservers
	}

	return result, nil
}

func parseResolvConf() []string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	var servers []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}
	return servers
}
