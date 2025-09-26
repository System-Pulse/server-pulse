package security

import (
	"fmt"
	"os/exec"
	"strings"
)


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
