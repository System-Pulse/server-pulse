package security

import (
	"fmt"
	"os/exec"
	"strings"
)

func (sm *SecurityManager) checkSSHPasswordAuthentication() SecurityCheck {
	cmd := exec.Command("sudo", "-n", "sshd", "-T")
	output, err := cmd.Output()
	if err != nil {
		return SecurityCheck{
			Name:    "SSH Password Authentication",
			Status:  "Error",
			Details: fmt.Sprintf("Failed to check SSH config: %v", err),
		}
	}

	config := make(map[string]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			config[parts[0]] = parts[1]
		}
	}

	passwordAuth := strings.ToLower(config["passwordauthentication"])

	if passwordAuth == "yes" {
		return SecurityCheck{
			Name:    "SSH Password Authentication",
			Status:  "Enabled",
			Details: "SSH password authentication is enabled - consider disabling for better security",
		}
	}

	return SecurityCheck{
		Name:    "SSH Password Authentication",
		Status:  "Disabled",
		Details: "SSH password authentication is disabled",
	}
}
