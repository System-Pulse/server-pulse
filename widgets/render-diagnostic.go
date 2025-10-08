package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/widgets/auth"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/bubbles/table"
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
		currentView = m.renderDiagnosticLogs()
	}
	return currentView
}

func (m Model) renderDiagnosticSecurity() string {
	doc := strings.Builder{}

	// Title
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Security Checks"))
	doc.WriteString("\n\n")

	// Authentication section
	if m.Diagnostic.AuthState == model.AuthRequired || m.Diagnostic.AuthState == model.AuthInProgress {
		authMessage := auth.GetAuthMessage(int(m.Diagnostic.AuthState), m.Diagnostic.AuthMessage)
		authStyle := auth.GetAuthStyle(int(m.Diagnostic.AuthState))

		doc.WriteString(authStyle.Render(authMessage))
		doc.WriteString("\n\n")
		doc.WriteString(m.Diagnostic.AuthMessage)
		doc.WriteString("\n\n")
		if m.Diagnostic.AuthState == model.AuthRequired {
			doc.WriteString(auth.AuthPromptStyle.Render("Enter Password:"))
			doc.WriteString("\n")
			doc.WriteString(m.Diagnostic.Password.View())
			doc.WriteString("\n\n")
			doc.WriteString(auth.AuthInfoStyle.Render(auth.AuthInstructions))
		} else {
			doc.WriteString(auth.AuthInProgressStyle.Render("⏳ " + m.Diagnostic.AuthMessage))
		}
		doc.WriteString("\n\n")
		return vars.CardStyle.Render(doc.String())
	}

	if m.Diagnostic.AuthState == model.AuthFailed {
		authMessage := auth.GetAuthMessage(int(m.Diagnostic.AuthState), m.Diagnostic.AuthMessage)
		authStyle := auth.GetAuthStyle(int(m.Diagnostic.AuthState))

		doc.WriteString(authStyle.Render(authMessage))
		doc.WriteString("\n\n")
		doc.WriteString(m.Diagnostic.AuthMessage)
		doc.WriteString("\n\n")
		doc.WriteString(auth.AuthInfoStyle.Render(auth.AuthRetryMessage))
		doc.WriteString("\n\n")
	}

	if m.Diagnostic.AuthState == model.AuthSuccess && m.Diagnostic.AuthTimer > 0 {
		authMessage := auth.GetAuthMessage(int(m.Diagnostic.AuthState), m.Diagnostic.AuthMessage)
		authStyle := auth.GetAuthStyle(int(m.Diagnostic.AuthState))

		doc.WriteString(authStyle.Render(authMessage))
		doc.WriteString("\n\n")
		doc.WriteString(auth.AuthInfoStyle.Render("Admin privileges granted"))
		doc.WriteString("\n\n")
	}

	if len(m.Diagnostic.SecurityChecks) == 0 {
		doc.WriteString("Loading security checks...")
		doc.WriteString("\n\n")
		return vars.CardStyle.Render(doc.String())
	}

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

	// Security table with access indicators
	var filteredRows []table.Row
	for _, check := range m.Diagnostic.SecurityChecks {
		// Check if user has access to this diagnostic
		hasAccess := m.canAccessDiagnostic(check.Name)

		// Add status icons based on status
		statusWithIcon := m.getSecurityStatusIcon(check.Status) + " " + check.Status

		// Add access indicator
		accessIndicator := ""
		if !hasAccess {
			accessIndicator = " " + auth.LockedCheckIndicator
			statusWithIcon = auth.LockedCheckIndicator + " " + check.Status
		}

		filteredRows = append(filteredRows, table.Row{
			check.Name + accessIndicator,
			statusWithIcon,
			check.Details,
		})
	}

	// Update table with filtered rows
	m.Diagnostic.SecurityTable.SetRows(filteredRows)
	doc.WriteString(m.Diagnostic.SecurityTable.View())

	// Footer with authentication info
	doc.WriteString("\n\n")
	if !m.Diagnostic.IsRoot && !m.Diagnostic.CanRunSudo {
		doc.WriteString(auth.AccessIndicatorStyle.Render(auth.AdminAccessRequired))
	} else if m.Diagnostic.AuthState == model.AuthSuccess && m.Diagnostic.AuthTimer > 0 {
		doc.WriteString(auth.AuthSuccessStyle.Render(auth.AdminAccessGranted))
	}

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

	doc := strings.Builder{}

	// Title
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Opened Ports Details"))
	doc.WriteString("\n\n")

	doc.WriteString(m.Diagnostic.PortsTable.View())
	doc.WriteString("\n\n")

	return vars.CardStyle.Render(doc.String())
}

func (m Model) renderFirewallDetails() string {
	if m.Diagnostic.FirewallInfo == nil {
		return vars.CardStyle.Render("No firewall information available")
	}

	doc := strings.Builder{}

	// Title
	title := fmt.Sprintf("Firewall Details - %s", m.Diagnostic.FirewallInfo.FirewallType)
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(title))
	doc.WriteString("\n\n")

	// Status and details
	statusStyle := lipgloss.NewStyle().Bold(true)
	if m.Diagnostic.FirewallInfo.Status == "Active" {
		statusStyle = statusStyle.Foreground(lipgloss.Color("46")) // Green
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("196")) // Red
	}

	doc.WriteString(vars.MetricLabelStyle.Render("Status: "))
	doc.WriteString(statusStyle.Render(m.Diagnostic.FirewallInfo.Status))
	doc.WriteString("\n")
	doc.WriteString(vars.MetricLabelStyle.Render("Details: "))
	doc.WriteString(m.Diagnostic.FirewallInfo.Details)
	doc.WriteString("\n\n")

	// Rules summary
	if len(m.Diagnostic.FirewallInfo.Rules) > 0 {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Firewall Rules (%d total)", len(m.Diagnostic.FirewallInfo.Rules))))
		doc.WriteString("\n\n")
		doc.WriteString(m.Diagnostic.FirewallTable.View())
		doc.WriteString("\n\n")
	} else {
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("No firewall rules configured"))
		doc.WriteString("\n\n")
	}

	// Raw output section (collapsible view)
	if m.Diagnostic.FirewallInfo.RawOutput != "" {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Complete Firewall Configuration:"))
		doc.WriteString("\n")
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("(Full raw output from firewall command)"))
		doc.WriteString("\n\n")

		// Show raw output in a bordered box
		rawOutputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MaxWidth(100)

		doc.WriteString(rawOutputStyle.Render(m.Diagnostic.FirewallInfo.RawOutput))
		doc.WriteString("\n")
	}

	return vars.CardStyle.Render(doc.String())
}

func (m Model) renderAutoBanDetails() string {
	if m.Diagnostic.AutoBanInfo == nil {
		return vars.CardStyle.Render("No auto-ban information available")
	}

	doc := strings.Builder{}

	// Title
	title := fmt.Sprintf("Auto Ban Details - %s", m.Diagnostic.AutoBanInfo.ServiceType)
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(title))
	doc.WriteString("\n\n")

	// Status and details
	statusStyle := lipgloss.NewStyle().Bold(true)
	if m.Diagnostic.AutoBanInfo.Status == "Active" {
		statusStyle = statusStyle.Foreground(lipgloss.Color("46")) // Green
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("196")) // Red
	}

	doc.WriteString(vars.MetricLabelStyle.Render("Status: "))
	doc.WriteString(statusStyle.Render(m.Diagnostic.AutoBanInfo.Status))
	doc.WriteString("\n")

	if m.Diagnostic.AutoBanInfo.Version != "" {
		doc.WriteString(vars.MetricLabelStyle.Render("Version: "))
		doc.WriteString(m.Diagnostic.AutoBanInfo.Version)
		doc.WriteString("\n")
	}

	doc.WriteString(vars.MetricLabelStyle.Render("Details: "))
	doc.WriteString(m.Diagnostic.AutoBanInfo.Details)
	doc.WriteString("\n\n")

	// Jails/Services table
	if len(m.Diagnostic.AutoBanInfo.Jails) > 0 {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Jails/Services (%d total)", len(m.Diagnostic.AutoBanInfo.Jails))))
		doc.WriteString("\n\n")
		doc.WriteString(m.Diagnostic.AutoBanTable.View())
		doc.WriteString("\n\n")
	} else {
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("No jails/services configured"))
		doc.WriteString("\n\n")
	}

	// Raw output section
	if m.Diagnostic.AutoBanInfo.RawOutput != "" {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Complete Configuration:"))
		doc.WriteString("\n")
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("(Full raw output from service)"))
		doc.WriteString("\n\n")

		rawOutputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MaxWidth(100)

		doc.WriteString(rawOutputStyle.Render(m.Diagnostic.AutoBanInfo.RawOutput))
		doc.WriteString("\n")
	}

	return vars.CardStyle.Render(doc.String())
}

func (m Model) renderDiagnosticLogs() string {
	doc := strings.Builder{}

	// Custom time input mode - show popup
	if m.Diagnostic.CustomTimeInputMode {
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Enter Custom Time Range"))
		doc.WriteString("\n")
		doc.WriteString(m.Diagnostic.LogTimeRangeInput.View())
		doc.WriteString("\n\n")

		// Show error message if validation failed
		if m.Diagnostic.CustomTimeInputError != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
			doc.WriteString(errorStyle.Render("✗ " + m.Diagnostic.CustomTimeInputError))
			doc.WriteString("\n\n")
		}

		doc.WriteString(lipgloss.NewStyle().Faint(true).Render("Examples: '2 hours ago', '2025-01-08 14:30:00', '3 days ago'"))
		doc.WriteString("\n")
		doc.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter to apply, ESC to cancel"))
		return vars.CardStyle.Render(doc.String())
	}

	// Filters section
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		MarginBottom(1)

	filterDoc := strings.Builder{}

	// Time range options
	timeRanges := []string{"All", "5m", "1h", "24h", "7d", "Custom"}
	timeRangeStr := "Time: "
	for i, tr := range timeRanges {
		if i == m.Diagnostic.LogTimeSelected {
			timeRangeStr += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).Render("["+tr+"]") + " "
		} else {
			timeRangeStr += tr + " "
		}
	}
	filterDoc.WriteString(timeRangeStr)
	filterDoc.WriteString("\n")

	// Log level options
	levels := []string{"All", "Error", "Warn", "Info", "Debug"}
	levelStr := "Level: "
	for i, level := range levels {
		if i == m.Diagnostic.LogLevelSelected {
			levelStr += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).Render("["+level+"]") + " "
		} else {
			levelStr += level + " "
		}
	}
	filterDoc.WriteString(levelStr)
	filterDoc.WriteString("\n")

	// Filter inputs
	filterDoc.WriteString("Search: " + m.Diagnostic.LogSearchInput.View() + "  ")
	filterDoc.WriteString("Service: " + m.Diagnostic.LogServiceInput.View())

	doc.WriteString(filterStyle.Render(filterDoc.String()))
	doc.WriteString("\n\n")

	// Log entries table or loading message
	if m.Diagnostic.LogsInfo == nil {
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Press 'r' to load logs or Enter to apply filters"))
		doc.WriteString("\n")
	} else if m.Diagnostic.LogsInfo.ErrorMsg != "" {
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: " + m.Diagnostic.LogsInfo.ErrorMsg))
		doc.WriteString("\n")
	} else if len(m.Diagnostic.LogsInfo.Entries) == 0 {
		doc.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("No log entries found matching filters"))
		doc.WriteString("\n")
	} else {
		// Display logs table
		doc.WriteString(m.Diagnostic.LogsTable.View())
		doc.WriteString("\n\n")

		// Summary
		summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
		summary := fmt.Sprintf("Showing %d entries", m.Diagnostic.LogsInfo.TotalCount)
		if m.Diagnostic.LogsInfo.HasMore {
			summary += " (more available - increase limit or refine filters)"
		}
		doc.WriteString(summaryStyle.Render(summary))
		doc.WriteString("\n")
	}

	return vars.CardStyle.Render(doc.String())
}
