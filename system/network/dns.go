package network

import (
	"bufio"
	"fmt"
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
			line := scanner.Text()
			if strings.HasPrefix(strings.TrimSpace(line), "nameserver") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					dnsServers = append(dnsServers, DNSInfo{Server: fields[1]})
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to read resolv.conf: %w", err))
		}

		return DNSMsg(dnsServers)
	}
}
