package security

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func NewSecurityManager() *SecurityManager {
	return &SecurityManager{}
}

type SecurityCheck struct {
	Name    string
	Status  string
	Details string
}

type SecurityCheckResult struct {
	Checks     []SecurityCheck
	LastUpdate time.Time
}

type SecurityMsg []SecurityCheck

func (sm *SecurityManager) RunSecurityChecks() tea.Cmd {
	return func() tea.Msg {
		checks := []SecurityCheck{
			sm.checkSSLCertificate(),
			sm.checkSSHRootLogin(),
			sm.checkOpenPorts(),
			sm.checkFirewallStatus(),
			sm.checkSystemUpdates(),
		}
		return SecurityMsg(checks)
	}
}
