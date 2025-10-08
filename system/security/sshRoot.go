package security

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SSHSystemChecker retrieves the active SSH daemon configuration.
type SSHSystemChecker struct{}

type SSHRootInfos struct {
	Status  string
	Details string
}

type SSHRootMsg SSHRootInfos

func NewSSHSystemChecker() *SSHSystemChecker {
	return &SSHSystemChecker{}
}

func (s *SSHSystemChecker) GetActiveConfig(sm *SecurityManager) (map[string]string, error) {
	var cmd *exec.Cmd
	var output []byte
	var err error

	// Use sudo with password if authenticated
	if sm != nil && sm.CanUseSudo && !sm.IsRoot && sm.SudoPassword != "" {
		cmd = exec.Command("sudo", "-S", "sshd", "-T")
		cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		output, err = cmd.Output()
	} else {
		// First, try running `sshd -T` without sudo.
		cmd = exec.Command("sshd", "-T")
		output, err = cmd.Output()

		// If the command fails, try again with `sudo -n`.
		if err != nil {
			cmd = exec.Command("sudo", "-n", "sshd", "-T")
			var errSudo error
			output, errSudo = cmd.Output()
			if errSudo != nil {
				return nil, fmt.Errorf("failed to execute 'sshd -T' (err: %v) and 'sudo -n sshd -T' (err: %v)", err, errSudo)
			}
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
			config[strings.ToLower(parts[0])] = parts[1]
		}
	}

	return config, nil
}

func (sm *SecurityManager) checkSSHRootLogin() SecurityCheck {
	s := NewSSHSystemChecker()

	config, err := s.GetActiveConfig(sm)
	if err != nil {
		return SecurityCheck{
			Name:    "SSH Root Login",
			Status:  "Error",
			Details: fmt.Sprintf("Failed to get SSH config: %v", err),
		}
	}

	// Check for the permitrootlogin key. If not present, the default is 'prohibit-password'.
	permitRootLogin, ok := config["permitrootlogin"]
	if !ok {
		permitRootLogin = "prohibit-password"
	}

	status := "Disabled"
	details := "SSH root login is disabled."

	switch strings.ToLower(permitRootLogin) {
	case "yes":
		status = "Enabled (with password)"
		details = "Root login with a password is permitted. This is a high security risk."
	case "without-password", "prohibit-password":
		status = "Enabled (key-only)"
		details = "Root can login, but only with key-based authentication."
	case "forced-commands-only":
		status = "Enabled (commands-only)"
		details = "Root can login via public key, but only to run specific commands."
	case "no":
	}

	return SecurityCheck{
		Name:    "SSH Root Login",
		Status:  status,
		Details: details,
	}
}

func (sm *SecurityManager) DisplaySSHRootInfos() tea.Cmd {
	return func() tea.Msg {
		check := sm.checkSSHRootLogin()
		return SSHRootMsg(SSHRootInfos{
			Status:  check.Status,
			Details: check.Details,
		})
	}
}
