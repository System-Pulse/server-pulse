package widgets

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// check if as root
func (m Model) isRoot() bool {
	return os.Geteuid() == 0
}

// check if sudo is available
func (m Model) isSudoAvailable() bool {
	_, err := exec.LookPath("sudo")
	return err == nil
}

// setRoot sets the root password
func (m Model) setRoot() error {
	if !m.isSudoAvailable() {
		return fmt.Errorf("sudo is not available on this system")
	}

	cm := exec.Command("sudo", "-S", "ls")
	cm.Stdin = strings.NewReader(m.Diagnostic.Password.Value() + "\n")
	cm.Stdout = nil
	cm.Stderr = nil
	err := cm.Run()
	if err != nil {
		return fmt.Errorf("failed to run command, invalid password or insufficient privileges")
	}

	return nil
}

// check if user can run sudo
func (m Model) canRunSudo() bool {
	if !m.isSudoAvailable() {
		return false
	}

	cmd := exec.Command("sudo", "-n", "ls")
	err := cmd.Run()
	return err == nil
}

// getAdminRequiredChecks returns the list of diagnostic checks that require admin privileges
func (m Model) getAdminRequiredChecks() []string {
	return []string{
		"Open Ports",
		"SSH Root Login",
		"Firewall Status",
		"System Updates",
	}
}

// canAccessDiagnostic checks if user can access a specific diagnostic check
func (m Model) canAccessDiagnostic(checkName string) bool {
	adminChecks := m.getAdminRequiredChecks()

	// Check if this diagnostic requires admin privileges
	if slices.Contains(adminChecks, checkName) {
		return m.isRoot() || m.canRunSudo()
	}

	// Non-admin checks are always accessible
	return true
}
