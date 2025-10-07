package security

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type AutoBanJail struct {
	Name        string
	Status      string
	Filter      string
	Actions     string
	CurrentBans int
	TotalBans   int
	Details     string
}

type AutoBanInfos struct {
	ServiceType string   // fail2ban, crowdsec, etc.
	Status      string   // Active, Disabled
	Version     string   // Version info
	Jails       []AutoBanJail
	BannedIPs   []string
	Details     string
	RawOutput   string
}

type AutoBanMsg AutoBanInfos

func (sm *SecurityManager) checkAutoBan() SecurityCheck {
	// Vérifier d'abord les services systemd
	if check := sm.checkSystemdServices(); check.Status != "" {
		return check
	}

	// Vérifier les processus en cours d'exécution
	if check := sm.checkRunningProcesses(); check.Status != "" {
		return check
	}

	return SecurityCheck{
		Name:    "Auto Ban",
		Status:  "Disabled",
		Details: "No intrusion prevention service detected (fail2ban, crowdsec, ossec, suricata, snort, denyhosts, sshguard)",
	}
}

func (sm *SecurityManager) checkSystemdServices() SecurityCheck {
	services := sm.getServiceDefinitions()

	for _, service := range services {
		cmd := exec.Command(service.command[0], service.command[1:]...)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		status := strings.TrimSpace(string(output))
		if status == "active" {
			details := service.detailsFunc()
			return SecurityCheck{
				Name:    "Auto Ban",
				Status:  "Enabled",
				Details: details,
			}
		}

		// Cas spécial pour OSSEC
		if service.name == "ossec" && status != "active" {
			cmd := exec.Command("systemctl", "is-active", "ossec-hids")
			if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "active" {
				return SecurityCheck{
					Name:    "Auto Ban",
					Status:  "Enabled",
					Details: "ossec-hids is active with real-time monitoring",
				}
			}
		}
	}

	return SecurityCheck{}
}

func (sm *SecurityManager) getServiceDefinitions() []struct {
	name        string
	command     []string
	detailsFunc func() string
} {
	return []struct {
		name        string
		command     []string
		detailsFunc func() string
	}{
		{
			name:    "fail2ban",
			command: []string{"systemctl", "is-active", "fail2ban"},
			detailsFunc: func() string {
				cmd := exec.Command("fail2ban-client", "status")
				if output, err := cmd.Output(); err == nil {
					outputStr := string(output)

					if strings.Contains(outputStr, "Jail list:") {
						jailsLine := strings.Split(outputStr, "Jail list:")[1]
						jails := strings.Split(strings.TrimSpace(jailsLine), ",")
						activeJails := 0
						for _, jail := range jails {
							if strings.TrimSpace(jail) != "" {
								activeJails++
							}
						}
						return fmt.Sprintf("fail2ban is active with %d configured jails", activeJails)
					}
				}
				return "fail2ban is active"
			},
		},
		{
			name:    "denyhosts",
			command: []string{"systemctl", "is-active", "denyhosts"},
			detailsFunc: func() string {
				return "denyhosts is enabled and active"
			},
		},
		{
			name:    "sshguard",
			command: []string{"systemctl", "is-active", "sshguard"},
			detailsFunc: func() string {
				return "sshguard is enabled and active"
			},
		},
		{
			name:    "crowdsec",
			command: []string{"systemctl", "is-active", "crowdsec"},
			detailsFunc: func() string {
				cmd := exec.Command("cscli", "metrics")
				if _, err := cmd.Output(); err == nil {
					return "crowdsec is active with behavioral analysis"
				}
				return "crowdsec is enabled and active"
			},
		},
		{
			name:    "ossec",
			command: []string{"systemctl", "is-active", "ossec"},
			detailsFunc: func() string {
				cmd := exec.Command("systemctl", "is-active", "ossec-hids")
				if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) == "active" {
					return "ossec-hids is active with real-time monitoring"
				}
				return "ossec is enabled and active"
			},
		},
		{
			name:    "suricata",
			command: []string{"systemctl", "is-active", "suricata"},
			detailsFunc: func() string {
				cmd := exec.Command("suricata", "--build-info")
				if output, err := cmd.Output(); err == nil && strings.Contains(string(output), "Suricata") {
					return "suricata is active with network intrusion detection"
				}
				return "suricata is enabled and active"
			},
		},
		{
			name:    "snort",
			command: []string{"systemctl", "is-active", "snort"},
			detailsFunc: func() string {
				cmd := exec.Command("snort", "-V")
				if output, err := cmd.Output(); err == nil && strings.Contains(string(output), "Snort") {
					return "snort is active with network intrusion detection"
				}
				return "snort is enabled and active"
			},
		},
	}
}

func (sm *SecurityManager) checkRunningProcesses() SecurityCheck {
	additionalChecks := []struct {
		name    string
		command []string
		process string
	}{
		{"crowdsec", []string{"pgrep", "crowdsec"}, "crowdsec"},
		{"ossec", []string{"pgrep", "ossec"}, "ossec"},
		{"suricata", []string{"pgrep", "suricata"}, "suricata"},
		{"snort", []string{"pgrep", "snort"}, "snort"},
	}

	var activeServices []string

	for _, check := range additionalChecks {
		cmd := exec.Command(check.command[0], check.command[1:]...)
		if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
			activeServices = append(activeServices, check.name)
		}
	}

	if len(activeServices) > 0 {
		return SecurityCheck{
			Name:    "Auto Ban",
			Status:  "Enabled",
			Details: fmt.Sprintf("Active intrusion prevention: %s", strings.Join(activeServices, ", ")),
		}
	}

	return SecurityCheck{}
}

// DisplayAutoBanInfos retrieves detailed auto-ban/intrusion prevention information
func (sm *SecurityManager) DisplayAutoBanInfos() tea.Cmd {
	return func() tea.Msg {
		// Check which service is active
		services := sm.getServiceDefinitions()

		for _, service := range services {
			cmd := exec.Command(service.command[0], service.command[1:]...)
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			status := strings.TrimSpace(string(output))
			if status == "active" {
				// Get detailed info for this service
				switch service.name {
				case "fail2ban":
					return AutoBanMsg(sm.getFail2banDetails())
				case "crowdsec":
					return AutoBanMsg(sm.getCrowdsecDetails())
				case "ossec":
					return AutoBanMsg(sm.getOSSECDetails())
				case "denyhosts":
					return AutoBanMsg(sm.getDenyhostsDetails())
				case "sshguard":
					return AutoBanMsg(sm.getSSHGuardDetails())
				case "suricata":
					return AutoBanMsg(sm.getSuricataDetails())
				case "snort":
					return AutoBanMsg(sm.getSnortDetails())
				}
			}
		}

		// No active service found
		return AutoBanMsg(AutoBanInfos{
			ServiceType: "None",
			Status:      "Disabled",
			Details:     "No intrusion prevention service detected",
			Jails:       []AutoBanJail{},
			BannedIPs:   []string{},
			RawOutput:   "",
		})
	}
}

// getFail2banDetails retrieves detailed fail2ban information
func (sm *SecurityManager) getFail2banDetails() AutoBanInfos {
	var cmd *exec.Cmd

	// Get fail2ban status
	if sm.CanUseSudo && !sm.IsRoot {
		cmd = exec.Command("sudo", "-S", "fail2ban-client", "status")
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command("fail2ban-client", "status")
	}

	output, err := cmd.Output()
	if err != nil {
		return AutoBanInfos{
			ServiceType: "fail2ban",
			Status:      "Active",
			Details:     "fail2ban is running but unable to get details",
			Jails:       []AutoBanJail{},
			BannedIPs:   []string{},
		}
	}

	rawOutput := string(output)
	jails := []AutoBanJail{}
	bannedIPs := []string{}

	// Parse jails list
	if strings.Contains(rawOutput, "Jail list:") {
		jailsLine := strings.Split(rawOutput, "Jail list:")[1]
		jailNames := strings.Split(strings.TrimSpace(jailsLine), ",")

		for _, jailName := range jailNames {
			jailName = strings.TrimSpace(jailName)
			if jailName == "" {
				continue
			}

			// Get detailed info for this jail
			if sm.CanUseSudo && !sm.IsRoot {
				cmd = exec.Command("sudo", "-S", "fail2ban-client", "status", jailName)
				if sm.SudoPassword != "" {
					cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
				}
			} else {
				cmd = exec.Command("fail2ban-client", "status", jailName)
			}

			jailOutput, err := cmd.Output()
			if err != nil {
				continue
			}

			jail := parseFailfbanJail(jailName, string(jailOutput))
			jails = append(jails, jail)

			// Collect banned IPs
			bannedIPs = append(bannedIPs, jail.Details)
		}
	}

	// Get version
	versionCmd := exec.Command("fail2ban-client", "version")
	version := ""
	if versionOutput, err := versionCmd.Output(); err == nil {
		version = strings.TrimSpace(string(versionOutput))
	}

	return AutoBanInfos{
		ServiceType: "fail2ban",
		Status:      "Active",
		Version:     version,
		Jails:       jails,
		BannedIPs:   bannedIPs,
		Details:     fmt.Sprintf("fail2ban is active with %d configured jails", len(jails)),
		RawOutput:   rawOutput,
	}
}

func parseFailfbanJail(name string, output string) AutoBanJail {
	jail := AutoBanJail{
		Name:   name,
		Status: "Active",
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Filter") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				jail.Filter = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Actions") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				jail.Actions = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Currently banned:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &jail.CurrentBans)
			}
		} else if strings.Contains(line, "Total banned:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &jail.TotalBans)
			}
		} else if strings.HasPrefix(line, "Banned IP list:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				jail.Details = strings.TrimSpace(parts[1])
			}
		}
	}

	if jail.Details == "" {
		jail.Details = fmt.Sprintf("Currently: %d banned, Total: %d banned", jail.CurrentBans, jail.TotalBans)
	}

	return jail
}

// getCrowdsecDetails retrieves CrowdSec information
func (sm *SecurityManager) getCrowdsecDetails() AutoBanInfos {
	// Get basic status
	cmd := exec.Command("cscli", "metrics")
	output, _ := cmd.Output()

	// Get version
	versionCmd := exec.Command("cscli", "version")
	version := ""
	if versionOutput, err := versionCmd.Output(); err == nil {
		version = strings.TrimSpace(strings.Split(string(versionOutput), "\n")[0])
	}

	return AutoBanInfos{
		ServiceType: "CrowdSec",
		Status:      "Active",
		Version:     version,
		Details:     "CrowdSec is active with behavioral analysis",
		Jails:       []AutoBanJail{},
		BannedIPs:   []string{},
		RawOutput:   string(output),
	}
}

// getOSSECDetails retrieves OSSEC information
func (sm *SecurityManager) getOSSECDetails() AutoBanInfos {
	return AutoBanInfos{
		ServiceType: "OSSEC",
		Status:      "Active",
		Details:     "OSSEC HIDS is active with real-time monitoring",
		Jails:       []AutoBanJail{},
		BannedIPs:   []string{},
		RawOutput:   "OSSEC is a host-based intrusion detection system",
	}
}

// getDenyhostsDetails retrieves DenyHosts information
func (sm *SecurityManager) getDenyhostsDetails() AutoBanInfos {
	return AutoBanInfos{
		ServiceType: "DenyHosts",
		Status:      "Active",
		Details:     "DenyHosts is enabled and active",
		Jails:       []AutoBanJail{},
		BannedIPs:   []string{},
		RawOutput:   "DenyHosts monitors SSH login attempts",
	}
}

// getSSHGuardDetails retrieves SSHGuard information
func (sm *SecurityManager) getSSHGuardDetails() AutoBanInfos {
	return AutoBanInfos{
		ServiceType: "SSHGuard",
		Status:      "Active",
		Details:     "SSHGuard is enabled and active",
		Jails:       []AutoBanJail{},
		BannedIPs:   []string{},
		RawOutput:   "SSHGuard protects SSH and other services",
	}
}

// getSuricataDetails retrieves Suricata information
func (sm *SecurityManager) getSuricataDetails() AutoBanInfos {
	// Get version
	cmd := exec.Command("suricata", "--build-info")
	output, _ := cmd.Output()

	version := ""
	if strings.Contains(string(output), "Suricata") {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			version = strings.TrimSpace(lines[0])
		}
	}

	return AutoBanInfos{
		ServiceType: "Suricata",
		Status:      "Active",
		Version:     version,
		Details:     "Suricata IDS/IPS is active with network intrusion detection",
		Jails:       []AutoBanJail{},
		BannedIPs:   []string{},
		RawOutput:   string(output),
	}
}

// getSnortDetails retrieves Snort information
func (sm *SecurityManager) getSnortDetails() AutoBanInfos {
	cmd := exec.Command("snort", "-V")
	output, _ := cmd.Output()

	version := ""
	if strings.Contains(string(output), "Snort") {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			version = strings.TrimSpace(lines[0])
		}
	}

	return AutoBanInfos{
		ServiceType: "Snort",
		Status:      "Active",
		Version:     version,
		Details:     "Snort IDS/IPS is active with network intrusion detection",
		Jails:       []AutoBanJail{},
		BannedIPs:   []string{},
		RawOutput:   string(output),
	}
}
