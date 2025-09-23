package widgets

import (
	"fmt"
	"strings"

	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderDignostics() string {
	currentView := ""
	switch m.Diagnostic.SelectedItem {
	case model.DiagnosticSecurityChecks:
		currentView = m.renderDiagnosticSecurity()
	case model.DiagnosticTabPerformances:
		currentView = renderNotImplemented("Performance Analysis")
	case model.DiagnosticTabLogs:
		currentView = renderNotImplemented("Log Analysis")
	}
	return currentView
}

func (m Model) renderDiagnosticSecurity() string {
	if len(m.Diagnostic.SecurityChecks) == 0 {
		return vars.CardStyle.Render("Loading security checks...")
	}

	doc := strings.Builder{}

	// Title
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Security Checks"))
	doc.WriteString("\n\n")

	// Domain input section
	if m.Diagnostic.DomainInputMode {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Enter Domain Name for SSL Check"))
		doc.WriteString("\n")
		doc.WriteString(m.Diagnostic.DomainInput.View())
		doc.WriteString("\n\n")
		doc.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter to check SSL, Esc to cancel"))
		doc.WriteString("\n\n")
	} else {
		// Show current domain or prompt to enter one
		currentDomain := m.Diagnostic.DomainInput.Value()
		if currentDomain == "" {
			doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Press 'd' to enter domain name for SSL check"))
		} else {
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Current Domain: ") + currentDomain)
			doc.WriteString("\n")
			doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Press 'd' to change domain, 'r' to refresh checks"))
		}
		doc.WriteString("\n\n")
	}

	// Security table
	doc.WriteString(m.Diagnostic.SecurityTable.View())

	// Last update info - we'll use current time for now since we don't track LastUpdate yet
	doc.WriteString("\n\n")

	return vars.CardStyle.Render(doc.String())
}

func (m Model) renderCertificateDetails() string {
	if m.Diagnostic.CertificateInfo == nil {
		return vars.CardStyle.Render("No certificate information available")
	}

	cert := m.Diagnostic.CertificateInfo
	doc := strings.Builder{}

	// Title
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("SSL Certificate Details"))
	doc.WriteString("\n\n")

	// Certificate information
	doc.WriteString(vars.MetricLabelStyle.Render("Subject: ") + cert.Subject + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Issuer: ") + cert.Issuer + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Serial Number: ") + cert.SerialNumber + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Version: ") + fmt.Sprintf("%d", cert.Version) + "\n")
	doc.WriteString("\n")

	// Validity period
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Validity Period"))
	doc.WriteString("\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Valid From: ") + cert.ValidityPeriodFrom + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Valid To: ") + cert.ValidityPeriodTo + "\n")

	// Days until expiry with color coding
	expiryText := ""
	if cert.DaysUntilExpiry < 30 {
		expiryText = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("⚠ " + fmt.Sprintf("%d", cert.DaysUntilExpiry) + " days (Expiring Soon!)")
	} else if cert.DaysUntilExpiry < 90 {
		expiryText = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⚠ " + fmt.Sprintf("%d", cert.DaysUntilExpiry) + " days")
	} else {
		expiryText = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("✓ " + fmt.Sprintf("%d", cert.DaysUntilExpiry) + " days")
	}
	doc.WriteString(vars.MetricLabelStyle.Render("Days Until Expiry: ") + expiryText + "\n")
	doc.WriteString("\n")

	// Security information
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Security Information"))
	doc.WriteString("\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Algorithm: ") + cert.Algorithm + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Signature Algorithm: ") + cert.SignatureAlgorithm + "\n")

	hostnameStatus := ""
	if cert.HostnameVerified {
		hostnameStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("✓ Verified")
	} else {
		hostnameStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗ Not Verified")
	}
	doc.WriteString(vars.MetricLabelStyle.Render("Hostname Verified: ") + hostnameStatus + "\n")
	doc.WriteString("\n")

	// Alternative names
	if len(cert.AlternativeNames) > 0 {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Alternative Names"))
		doc.WriteString("\n")
		for _, name := range cert.AlternativeNames {
			doc.WriteString("• " + name + "\n")
		}
		doc.WriteString("\n")
	}

	return vars.CardStyle.Render(doc.String())
}

func (m Model) renderSSHRootDetails() string {
	if m.Diagnostic.SSHRootInfo == nil {
		return vars.CardStyle.Render("No SSH root login information available")
	}

	sshInfo := m.Diagnostic.SSHRootInfo
	doc := strings.Builder{}

	// Title
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("SSH Root Login Details"))
	doc.WriteString("\n\n")

	doc.WriteString(vars.MetricLabelStyle.Render("Status: ") + sshInfo.Status + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Details: ") + sshInfo.Details + "\n")

	return vars.CardStyle.Render(doc.String())
}

func (m Model) renderOpenedPortsDetails() string {
	if m.Diagnostic.OpenedPortsInfo == nil {
		return vars.CardStyle.Render("No opened ports information available")
	}

	openedPorts := m.Diagnostic.OpenedPortsInfo
	doc := strings.Builder{}

	// Title
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Opened Ports Details"))
	doc.WriteString("\n\n")

	// Opened ports information
	doc.WriteString(vars.MetricLabelStyle.Render("Total Opened Ports: ") + fmt.Sprintf("%d", len(openedPorts.Ports)) + "\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Ports: ") + showPorts(openedPorts.Ports) + "\n")

	return vars.CardStyle.Render(doc.String())
}

func showPorts(ports []int) string {
	if len(ports) == 0 {
		return "No opened ports"
	}

	var portList []string
	for _, port := range ports {
		portList = append(portList, fmt.Sprintf("%d", port))
	}
	return strings.Join(portList, ", ")
}
