package widgets

import (
	"fmt"
	v "github.com/System-Pulse/server-pulse/widgets/vars"
	"math/rand"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// check if as root
func isRoot() bool {
	return os.Geteuid() == 0
}

// check if sudo is available
func isSudoAvailable() bool {
	_, err := exec.LookPath("sudo")
	return err == nil
}

// setRoot sets the root password
func (m Model) setRoot() error {
	if !isSudoAvailable() {
		return fmt.Errorf("sudo is not available on this system")
	}
	m.Diagnostic.SudoAvailable = true
	m.Diagnostic.CanRunSudo = true
	m.Network.SudoAvailable = true
	m.Network.CanRunSudo = true
	cm := exec.Command("sudo", "-S", "ls")
	cm.Stdin = strings.NewReader(m.Diagnostic.Password.Value() + "\n")
	cm.Stdout = nil
	cm.Stderr = nil
	err := cm.Run()
	if err != nil {
		return fmt.Errorf("failed to run command, invalid password or insufficient privileges")
	}
	m.Diagnostic.IsRoot = true
	m.Network.IsRoot = true
	return nil
}

// setNetworkRoot sets the root password for network operations
func (m Model) setNetworkRoot() error {
	if !isSudoAvailable() {
		return fmt.Errorf("sudo is not available on this system")
	}
	m.Network.SudoAvailable = true
	m.Network.CanRunSudo = true
	cm := exec.Command("sudo", "-S", "ls")
	cm.Stdin = strings.NewReader(m.Diagnostic.Password.Value() + "\n")
	cm.Stdout = nil
	cm.Stderr = nil
	err := cm.Run()
	if err != nil {
		return fmt.Errorf("failed to run command, invalid password or insufficient privileges")
	}
	m.Network.IsRoot = true
	return nil
}

// check if user can run sudo
func canRunSudo() bool {
	if !isSudoAvailable() {
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
		return m.Diagnostic.IsRoot || m.Diagnostic.CanRunSudo
	}

	// Non-admin checks are always accessible
	return true
}

// canAccessNetworkConnections checks if user can access detailed network connections
func (m Model) canAccessNetworkConnections() bool {
	return m.Network.IsRoot || m.Network.CanRunSudo
}

func getRandomSpinner() spinner.Model {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomSpinner := v.Spinners[r.Intn(len(v.Spinners))]

	colors := []lipgloss.Color{
		lipgloss.Color("205"), // Magenta
		lipgloss.Color("39"),  // Blue
		lipgloss.Color("46"),  // Green
		lipgloss.Color("214"), // Orange
		lipgloss.Color("198"), // Pink
		lipgloss.Color("51"),  // Cyan
		lipgloss.Color("226"), // Yellow
	}
	randomColor := colors[r.Intn(len(colors))]

	spinnerModel := spinner.New()
	spinnerModel.Spinner = randomSpinner
	spinnerModel.Style = lipgloss.NewStyle().Foreground(randomColor)

	return spinnerModel
}

func (m *Model) resetSpinner() {
	m.Ui.Spinner = getRandomSpinner()
}
