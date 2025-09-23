package security

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type OpenedPorts struct{}

type OpenedPortsInfos struct {
	Ports   []int
	Details string
	Status  string
}

type OpenedPortsMsg OpenedPortsInfos

func NewOpenedPortsChecker() *OpenedPorts {
	return &OpenedPorts{}
}

func (o *OpenedPorts) GetOpenedPorts() (map[int]bool, error) {
	cmd := exec.Command("ss", "-tlpn")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ss command: %w", err)
	}

	openPorts := make(map[int]bool)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if !strings.Contains(line, "LISTEN") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		localAddr := fields[3]
		portStr := extractPortFromAddr(localAddr)

		if portStr != "" {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				continue
			}
			openPorts[port] = true
		}
	}

	return openPorts, nil
}

func extractPortFromAddr(addr string) string {
	if lastColon := strings.LastIndex(addr, ":"); lastColon != -1 {
		return addr[lastColon+1:]
	}
	return ""
}

func (sm *SecurityManager) checkOpenPorts() SecurityCheck {
	o := NewOpenedPortsChecker()

	openPorts, err := o.GetOpenedPorts()
	if err != nil {
		return SecurityCheck{
			Name:    "Open Ports",
			Status:  "Error",
			Details: fmt.Sprintf("Failed to get open ports: %v", err),
		}
	}

	if len(openPorts) == 0 {
		return SecurityCheck{
			Name:    "Open Ports",
			Status:  "Secure",
			Details: "No open ports detected",
		}
	}

	for port := range openPorts {
		if port == 22 {
			return SecurityCheck{
				Name:    "Open Ports",
				Status:  "Warning",
				Details: "Port 22 (SSH) is open. Change default ssh port for more security.",
			}
		}
	}

	return SecurityCheck{
		Name:    "Open Ports",
		Status:  "Secure",
		Details: fmt.Sprintf("%d Opened ports: %s", len(openPorts), ShowOpenedPortsDetails(openPorts)),
	}
}

func ShowOpenedPortsDetails(ports map[int]bool) string {
	if len(ports) == 0 {
		return "No open ports detected"
	}

	var portList []string
	for port := range ports {
		portList = append(portList, fmt.Sprintf("%d", port))
	}
	return fmt.Sprintf("%s", strings.Join(portList, ", "))
}
