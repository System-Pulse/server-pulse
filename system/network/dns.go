package network

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func GetDNS() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open("/etc/resolv.conf")
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to open resolv.conf: %w", err))
		}
		defer file.Close()

		var dnsServers []DNSInfo
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			if strings.HasPrefix(line, "nameserver") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					server := fields[1]
					if net.ParseIP(server) != nil {
						dnsServers = append(dnsServers, DNSInfo{Server: server})
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to read resolv.conf: %w", err))
		}

		if len(dnsServers) == 0 {
			return DNSMsg([]DNSInfo{{Server: "No DNS servers found"}})
		}

		return DNSMsg(dnsServers)
	}
}
