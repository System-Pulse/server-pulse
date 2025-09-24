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

func (sm *SecurityManager) RunSecurityChecks(domain string) tea.Cmd {
	return func() tea.Msg {
		checks := []SecurityCheck{
			sm.checkSSLCertificate(domain),
			sm.checkSSHRootLogin(),
			sm.checkSSHPasswordAuthentication(),
			sm.checkPasswordPolicy(),
			sm.checkOpenPorts(),
			sm.checkFirewallStatus(),
			sm.checkAutoBan(),
			sm.checkSystemUpdates(),
			sm.checkSystemRestart(),
		}
		return SecurityMsg(checks)
	}
}
