package security

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type FirewallRule struct {
	Description string // Human-readable rule description
	RawRule     string // Raw rule text for reference
}

type FirewallInfos struct {
	FirewallType string
	Status       string
	Rules        []FirewallRule
	Details      string
	RawOutput    string // Complete raw firewall output for advanced users
}

type FirewallMsg FirewallInfos

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
	var cmd *exec.Cmd

	// Use sudo if authenticated and not running as root
	if sm.CanUseSudo && !sm.IsRoot {
		args := append([]string{"-S"}, fw.command...)
		cmd = exec.Command("sudo", args...)
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command(fw.command[0], fw.command[1:]...)
	}

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

// DisplayFirewallInfos retrieves detailed firewall rules and information
func (sm *SecurityManager) DisplayFirewallInfos() tea.Cmd {
	return func() tea.Msg {
		firewalls := sm.getFirewallConfigs()

		for _, fw := range firewalls {
			output, err := sm.executeFirewallCommand(fw)
			if err != nil {
				continue
			}

			// Check if this firewall is active
			check := sm.analyzeFirewallOutput(fw.name, output)
			if check != nil && check.Status == "Active" {
				// Get detailed rules for this firewall
				rules := sm.getFirewallRules(fw.name)

				// Get raw output for advanced view
				rawOutput := sm.getRawFirewallOutput(fw.name)

				return FirewallMsg(FirewallInfos{
					FirewallType: fw.name,
					Status:       check.Status,
					Rules:        rules,
					Details:      check.Details,
					RawOutput:    rawOutput,
				})
			}
		}

		// No active firewall found
		return FirewallMsg(FirewallInfos{
			FirewallType: "None",
			Status:       "Inactive",
			Rules:        []FirewallRule{},
			Details:      "No active firewall detected",
			RawOutput:    "",
		})
	}
}

// getRawFirewallOutput gets the complete raw output of firewall rules
func (sm *SecurityManager) getRawFirewallOutput(firewallType string) string {
	var cmd *exec.Cmd

	switch firewallType {
	case "UFW":
		if sm.CanUseSudo && !sm.IsRoot {
			cmd = exec.Command("sudo", "-S", "ufw", "status", "verbose")
			if sm.SudoPassword != "" {
				cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
			}
		} else {
			cmd = exec.Command("ufw", "status", "verbose")
		}
	case "firewalld":
		if sm.CanUseSudo && !sm.IsRoot {
			cmd = exec.Command("sudo", "-S", "firewall-cmd", "--list-all")
			if sm.SudoPassword != "" {
				cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
			}
		} else {
			cmd = exec.Command("firewall-cmd", "--list-all")
		}
	case "iptables":
		if sm.CanUseSudo && !sm.IsRoot {
			cmd = exec.Command("sudo", "-S", "iptables", "-L", "-v", "-n")
			if sm.SudoPassword != "" {
				cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
			}
		} else {
			cmd = exec.Command("iptables", "-L", "-v", "-n")
		}
	case "nftables":
		if sm.CanUseSudo && !sm.IsRoot {
			cmd = exec.Command("sudo", "-S", "nft", "list", "ruleset")
			if sm.SudoPassword != "" {
				cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
			}
		} else {
			cmd = exec.Command("nft", "list", "ruleset")
		}
	default:
		return "No firewall detected"
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error retrieving raw output: %v", err)
	}

	return string(output)
}

// getFirewallRules retrieves detailed rules based on firewall type
func (sm *SecurityManager) getFirewallRules(firewallType string) []FirewallRule {
	switch firewallType {
	case "UFW":
		return sm.getUFWRules()
	case "firewalld":
		return sm.getFirewalldRules()
	case "iptables":
		return sm.getIptablesRules()
	case "nftables":
		return sm.getNftablesRules()
	default:
		return []FirewallRule{}
	}
}

// getUFWRules retrieves UFW firewall rules
func (sm *SecurityManager) getUFWRules() []FirewallRule {
	var cmd *exec.Cmd
	if sm.CanUseSudo && !sm.IsRoot {
		cmd = exec.Command("sudo", "-S", "ufw", "status", "numbered")
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command("ufw", "status", "numbered")
	}

	output, err := cmd.Output()
	if err != nil {
		return []FirewallRule{}
	}

	rules := []FirewallRule{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Status:") || strings.HasPrefix(line, "To") || strings.HasPrefix(line, "--") {
			continue
		}

		// Parse UFW rule format: [ 1] 22/tcp ALLOW IN Anywhere
		if strings.HasPrefix(line, "[") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				ruleNum := strings.Trim(parts[0], "[]")
				port := parts[1]
				action := parts[2]
				direction := ""
				from := ""

				// Parse direction and source
				if len(parts) >= 5 {
					direction = parts[3]
				}
				if len(parts) >= 6 {
					from = strings.Join(parts[4:], " ")
				}

				description := fmt.Sprintf("[%s] %s %s %s from %s", ruleNum, action, direction, port, from)

				rule := FirewallRule{
					Description: description,
					RawRule:     line,
				}
				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// getFirewalldRules retrieves firewalld rules
func (sm *SecurityManager) getFirewalldRules() []FirewallRule {
	var cmd *exec.Cmd
	if sm.CanUseSudo && !sm.IsRoot {
		cmd = exec.Command("sudo", "-S", "firewall-cmd", "--list-all")
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command("firewall-cmd", "--list-all")
	}

	output, err := cmd.Output()
	if err != nil {
		return []FirewallRule{}
	}

	rules := []FirewallRule{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "services:") {
			services := strings.TrimPrefix(line, "services:")
			services = strings.TrimSpace(services)
			if services != "" {
				for _, svc := range strings.Fields(services) {
					description := fmt.Sprintf("Service: %s (ALLOWED)", svc)
					rules = append(rules, FirewallRule{
						Description: description,
						RawRule:     line,
					})
				}
			}
		} else if strings.HasPrefix(line, "ports:") {
			ports := strings.TrimPrefix(line, "ports:")
			ports = strings.TrimSpace(ports)
			if ports != "" {
				for _, port := range strings.Fields(ports) {
					description := fmt.Sprintf("Port: %s (ALLOWED)", port)
					rules = append(rules, FirewallRule{
						Description: description,
						RawRule:     line,
					})
				}
			}
		} else if strings.HasPrefix(line, "rich rules:") {
			// Capture rich rules if present
			rules = append(rules, FirewallRule{
				Description: "Rich rules configured (see raw output)",
				RawRule:     line,
			})
		}
	}

	return rules
}

// getIptablesRules retrieves iptables rules
func (sm *SecurityManager) getIptablesRules() []FirewallRule {
	var cmd *exec.Cmd
	if sm.CanUseSudo && !sm.IsRoot {
		cmd = exec.Command("sudo", "-S", "iptables", "-L", "-n", "--line-numbers")
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command("iptables", "-L", "-n", "--line-numbers")
	}

	output, err := cmd.Output()
	if err != nil {
		return []FirewallRule{}
	}

	rules := []FirewallRule{}
	lines := strings.Split(string(output), "\n")
	currentChain := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect chain header
		if strings.HasPrefix(line, "Chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentChain = parts[1]
			}
			continue
		}

		// Skip header lines
		if strings.HasPrefix(line, "num") || strings.HasPrefix(line, "target") {
			continue
		}

		// Parse rule line - format is variable, so we'll be flexible
		parts := strings.Fields(line)
		if len(parts) >= 4 && currentChain != "" {
			ruleNum := parts[0]
			target := parts[1]
			protocol := parts[2]

			// Build a descriptive text
			description := fmt.Sprintf("[%s] Chain: %s | Target: %s | Protocol: %s", ruleNum, currentChain, target, protocol)

			// Add source and destination if available
			if len(parts) >= 6 {
				src := parts[4]
				dst := parts[5]
				if src != "0.0.0.0/0" {
					description += fmt.Sprintf(" | From: %s", src)
				}
				if dst != "0.0.0.0/0" {
					description += fmt.Sprintf(" | To: %s", dst)
				}
			}

			// Add any additional info
			if len(parts) > 6 {
				extra := strings.Join(parts[6:], " ")
				if extra != "" {
					description += fmt.Sprintf(" | %s", extra)
				}
			}

			rules = append(rules, FirewallRule{
				Description: description,
				RawRule:     line,
			})
		}
	}

	return rules
}

// getNftablesRules retrieves nftables rules
func (sm *SecurityManager) getNftablesRules() []FirewallRule {
	var cmd *exec.Cmd
	if sm.CanUseSudo && !sm.IsRoot {
		cmd = exec.Command("sudo", "-S", "nft", "list", "ruleset")
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command("nft", "list", "ruleset")
	}

	output, err := cmd.Output()
	if err != nil {
		return []FirewallRule{}
	}

	rules := []FirewallRule{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for actual rule lines (they typically have accept/drop/reject)
		if strings.Contains(line, "accept") || strings.Contains(line, "drop") || strings.Contains(line, "reject") {
			action := "ACCEPT"
			if strings.Contains(line, "drop") {
				action = "DROP"
			} else if strings.Contains(line, "reject") {
				action = "REJECT"
			}

			// Create a more readable description
			description := fmt.Sprintf("%s | %s", action, line)

			rule := FirewallRule{
				Description: description,
				RawRule:     line,
			}
			rules = append(rules, rule)
		}
	}

	return rules
}
