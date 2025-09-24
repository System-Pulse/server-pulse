package security

import (
	"fmt"
	"os/exec"
	"strings"
)

func (sm *SecurityManager) checkFirewallStatus() SecurityCheck {

	firewalls := []struct {
		name    string
		command []string
	}{
		{"UFW", []string{"ufw", "status"}},
		{"firewalld", []string{"firewall-cmd", "--state"}},
		{"iptables", []string{"iptables", "-L"}},
		{"nftables", []string{"nft", "list ruleset"}},
	}

	for _, fw := range firewalls {
		cmd := exec.Command(fw.command[0], fw.command[1:]...)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		outputStr := strings.ToLower(string(output))

		switch fw.name {
		case "UFW":
			if strings.Contains(outputStr, "status: active") {
				return SecurityCheck{
					Name:    "Firewall Status",
					Status:  "Active",
					Details: fmt.Sprintf("%s is active and properly configured", fw.name),
				}
			} else if strings.Contains(outputStr, "status: inactive") {
				return SecurityCheck{
					Name:    "Firewall Status",
					Status:  "Inactive",
					Details: fmt.Sprintf("%s is installed but inactive", fw.name),
				}
			}

		case "firewalld":
			if strings.Contains(outputStr, "running") {
				return SecurityCheck{
					Name:    "Firewall Status",
					Status:  "Active",
					Details: fmt.Sprintf("%s is active and properly configured", fw.name),
				}
			}

		case "iptables":

			if len(strings.Split(string(output), "\n")) > 10 {
				return SecurityCheck{
					Name:    "Firewall Status",
					Status:  "Active",
					Details: fmt.Sprintf("%s rules are configured", fw.name),
				}
			}
		case "nftables":

			if len(strings.Split(string(output), "\n")) > 10 {
				return SecurityCheck{
					Name:    "Firewall Status",
					Status:  "Active",
					Details: fmt.Sprintf("%s rules are configured", fw.name),
				}
			}
		}

	}

	return SecurityCheck{
		Name:    "Firewall Status",
		Status:  "Inactive",
		Details: "No active firewall detected (UFW, firewalld, or iptables)",
	}
}
