package security

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

func (o *OpenedPorts) GetOpenedPorts(sm *SecurityManager) (map[int]bool, error) {
	var cmd *exec.Cmd

	// Use sudo if authenticated and not running as root
	if sm != nil && sm.CanUseSudo && !sm.IsRoot {
		cmd = exec.Command("sudo", "-S", "ss", "-tlpn")
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command("ss", "-tlpn")
	}

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

type RiskLevel int

const (
	Secure RiskLevel = iota
	Warning
	HighRisk
)

func (sm *SecurityManager) checkOpenPorts() SecurityCheck {
	o := NewOpenedPortsChecker()

	openPorts, err := o.GetOpenedPorts(sm)
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

	maxRisk := Secure
	var riskPorts []string
	var criticalFindings []string

	for port := range openPorts {
		risk, message := analyzePortRisk(port)

		if risk > maxRisk {
			maxRisk = risk
			if risk == HighRisk {
				criticalFindings = []string{message}
				riskPorts = []string{fmt.Sprintf("%d", port)}
			}
		} else if risk == maxRisk && risk == HighRisk {
			criticalFindings = append(criticalFindings, message)
			riskPorts = append(riskPorts, fmt.Sprintf("%d", port))
		}
	}
	return buildSecurityCheckResult(maxRisk, openPorts, riskPorts, criticalFindings)
}

func analyzePortRisk(port int) (RiskLevel, string) {
	switch port {
	case 22:
		return Warning, "Port 22 (SSH) is open. Change default ssh port for more security."
	case 20, 21:
		return Warning, "Port 20/21 (FTP) is open. FTP is insecure, consider using SFTP or FTPS."
	case 23:
		return HighRisk, "Port 23 (Telnet) is open. Telnet is insecure and should be closed."
	case 161, 162:
		return Warning, "Port 161/162 (SNMP) is open. SNMP can expose sensitive information."
	case 137, 138, 139:
		return HighRisk, "Port 137/138/139 (NetBIOS), you can be attacked by null sessions."
	case 445:
		return HighRisk, "Port 445 (SMB) is open. SMB has had many vulnerabilities."
	case 3389:
		return HighRisk, "Port 3389 (RDP) is open. RDP is often targeted by attackers."
	case 3306, 5432, 6379, 27017:
		return HighRisk, "Database port exposed. Databases should not be accessible from internet."
	default:
		return Secure, ""
	}
}

func buildSecurityCheckResult(maxRisk RiskLevel, allPorts map[int]bool, riskPorts []string, findings []string) SecurityCheck {
	switch maxRisk {
	case HighRisk:
		return SecurityCheck{
			Name:   "Open Ports",
			Status: "High Risk",
			Details: fmt.Sprintf("Critical ports detected (%s): %s",
				strings.Join(riskPorts, ", "), strings.Join(findings, " | ")),
		}
	case Warning:
		return SecurityCheck{
			Name:    "Open Ports",
			Status:  "Warning",
			Details: strings.Join(findings, " | "),
		}
	default:
		return SecurityCheck{
			Name:    "Open Ports",
			Status:  "Secure",
			Details: fmt.Sprintf("%d ports open: %s", len(allPorts), ShowOpenedPorts(allPorts)),
		}
	}
}

func ShowOpenedPorts(ports map[int]bool) string {
	if len(ports) == 0 {
		return "No open ports detected"
	}

	var portList []string
	for port := range ports {
		portList = append(portList, fmt.Sprintf("%d", port))
	}
	return fmt.Sprintf("%s", strings.Join(portList, ", "))
}

func (sm *SecurityManager) DisplayOpenedPortsInfos() tea.Cmd {
	o := NewOpenedPortsChecker()
	var AllPorts []int
	value, err := o.GetOpenedPorts(sm)
	if err != nil {
		return nil
	}
	for port, b := range value {
		if b {
			AllPorts = append(AllPorts, port)
		}
	}

	return func() tea.Msg {
		openedPorts := sm.checkOpenPorts()
		return OpenedPortsMsg(OpenedPortsInfos{
			Ports:   AllPorts,
			Details: openedPorts.Details,
			Status:  openedPorts.Status,
		})
	}
}
