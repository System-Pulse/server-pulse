package security

import (
	"fmt"
	"os/exec"
	"strings"
)

type SSHSystemChecker struct {
	configPath string
}

func NewSSHSystemChecker() *SSHSystemChecker {
	return &SSHSystemChecker{
		configPath: "/etc/ssh/sshd_config",
	}
}

func (s *SSHSystemChecker) GetActiveConfig() (map[string]string, error) {
	cmd := exec.Command("sshd", "-T")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute sshd -T: %v", err)
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

	return config, nil
}

func (sm *SecurityManager) checkSSHRootLogin() SecurityCheck {
	s := NewSSHSystemChecker()

	config, err := s.GetActiveConfig()
	if err != nil {
		return SecurityCheck{
			Name:    "SSH Root Login",
			Status:  "Error",
			Details: fmt.Sprintf("Failed to get SSH config: %v", err),
		}
	}

	permitRootLogin := strings.ToLower(config["permitrootlogin"])
	enabled := (permitRootLogin == "yes" ||
		permitRootLogin == "without-password" ||
		permitRootLogin == "prohibit-password")

	status := "Disabled"
	if enabled {
		status = "Enabled"
	}
	return SecurityCheck{
		Name:    "SSH Root Login",
		Status:  status,
		Details: "SSH root login is " + status,
	}
}