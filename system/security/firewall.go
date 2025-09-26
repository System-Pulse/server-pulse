package security

import (
	"os/exec"
	"strings"
)

func (sm *SecurityManager) checkFirewallStatus() SecurityCheck {
	firewalls := sm.getFirewallConfigs()

	for _, fw := range firewalls {
		output, err := sm.executeFirewallCommand(fw)
		if err != nil {
			continue
		}

		check := sm.analyzeFirewallOutput(fw.name, output)
		if check != nil {
			return *check
		}
	}

	return sm.createInactiveFirewallCheck()
}

func (sm *SecurityManager) getFirewallConfigs() []struct {
	name    string
	command []string
} {
	return []struct {
		name    string
		command []string
	}{
		{"UFW", []string{"ufw", "status"}},
		{"firewalld", []string{"firewall-cmd", "--state"}},
		{"iptables", []string{"iptables", "-L"}},
		{"nftables", []string{"nft", "list ruleset"}},
	}
}

func (sm *SecurityManager) executeFirewallCommand(fw struct {
	name    string
	command []string
}) ([]byte, error) {
	cmd := exec.Command(fw.command[0], fw.command[1:]...)
	return cmd.Output()
}

func (sm *SecurityManager) analyzeFirewallOutput(firewallName string, output []byte) *SecurityCheck {
	outputStr := strings.ToLower(string(output))

	switch firewallName {
	case "UFW":
		return sm.analyzeUFWStatus(outputStr)
	case "firewalld":
		return sm.analyzeFirewalldStatus(outputStr)
	case "iptables":
		return sm.analyzeIptablesStatus(output)
	case "nftables":
		return sm.analyzeNftablesStatus(output)
	}

	return nil
}

func (sm *SecurityManager) analyzeUFWStatus(outputStr string) *SecurityCheck {
	if strings.Contains(outputStr, "status: active") {
		return &SecurityCheck{
			Name:    "Firewall Status",
			Status:  "Active",
			Details: "UFW is active and properly configured",
		}
	} else if strings.Contains(outputStr, "status: inactive") {
		return &SecurityCheck{
			Name:    "Firewall Status",
			Status:  "Inactive",
			Details: "UFW is installed but inactive",
		}
	}
	return nil
}

func (sm *SecurityManager) analyzeFirewalldStatus(outputStr string) *SecurityCheck {
	if strings.Contains(outputStr, "running") {
		return &SecurityCheck{
			Name:    "Firewall Status",
			Status:  "Active",
			Details: "firewalld is active and properly configured",
		}
	}
	return nil
}

func (sm *SecurityManager) analyzeIptablesStatus(output []byte) *SecurityCheck {
	if len(strings.Split(string(output), "\n")) > 10 {
		return &SecurityCheck{
			Name:    "Firewall Status",
			Status:  "Active",
			Details: "iptables rules are configured",
		}
	}
	return nil
}

func (sm *SecurityManager) analyzeNftablesStatus(output []byte) *SecurityCheck {
	if len(strings.Split(string(output), "\n")) > 10 {
		return &SecurityCheck{
			Name:    "Firewall Status",
			Status:  "Active",
			Details: "nftables rules are configured",
		}
	}
	return nil
}

func (sm *SecurityManager) createInactiveFirewallCheck() SecurityCheck {
	return SecurityCheck{
		Name:    "Firewall Status",
		Status:  "Inactive",
		Details: "No active firewall detected (UFW, firewalld, or iptables)",
	}
}
