package widgets

import (
	"fmt"
	"strings"

	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/lipgloss"
)

// renderReporting renders the main reporting view
func (m Model) renderReporting() string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	content.WriteString(headerStyle.Render("📊 System Health Reports"))
	content.WriteString("\n\n")

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	content.WriteString(instructions.Render("Press 'g' to generate a new report, 's' to save current report, 'l' to load saved reports"))
	content.WriteString("\n\n")

	// Report status
	if m.Reporting.IsGenerating {
		content.WriteString(m.renderGeneratingReport())
	} else if m.Reporting.IsSaving {
		content.WriteString(m.renderSavingReport())
	} else if m.Ui.State == model.StateViewingReport {
		content.WriteString(m.renderViewingReport())
	} else if m.Reporting.ShowSavedReports {
		content.WriteString(m.renderSavedReports())
	} else {
		// Show save notification if available
		if m.Reporting.SaveNotification != "" {
			content.WriteString(m.renderSaveNotification())
		}
		content.WriteString(m.renderReportMenu())
	}

	return content.String()
}

// renderReportMenu renders the main reporting menu
func (m Model) renderReportMenu() string {
	var content strings.Builder

	// Last generated report info
	if !m.Reporting.LastGenerated.IsZero() {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Italic(true)

		content.WriteString(infoStyle.Render(fmt.Sprintf("Last report generated: %s",
			m.Reporting.LastGenerated.Format("2006-01-02 15:04:05"))))
		content.WriteString("\n\n")
	}

	// Available actions
	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Bold(true)

	content.WriteString(actionStyle.Render("Available Actions:"))
	content.WriteString("\n\n")

	actions := []struct {
		key         string
		description string
	}{
		{"g", "Generate new system health report"},
		{"s", "Save current report to file"},
		{"l", "Load and view saved reports"},
	}

	for _, action := range actions {
		content.WriteString(fmt.Sprintf("  %s - %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render(action.key),
			action.description))
	}

	return content.String()
}

// renderSavedReports renders the saved reports list
func (m Model) renderSavedReports() string {
	var content strings.Builder

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Bold(true)

	content.WriteString(actionStyle.Render("Saved Reports:"))
	content.WriteString("\n\n")

	m.Reporting.RefreshSavedReports()
	if len(m.Reporting.SavedReports) > 0 {
		for i, report := range m.Reporting.SavedReports {
			numberStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)

			reportStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

			if i == m.Reporting.SelectedReport {
				reportStyle = reportStyle.
					Background(lipgloss.Color("57")).
					Foreground(lipgloss.Color("229"))
			}

			content.WriteString(fmt.Sprintf("  %s %s\n",
				numberStyle.Render(fmt.Sprintf("%d.", i+1)),
				reportStyle.Render(report)))
		}

		content.WriteString("\n")
		content.WriteString(instructionsStyle().Render("Press ENTER to view selected report, 'd' to delete, 'b' to go back"))
	} else {
		content.WriteString(instructionsStyle().Render("No saved reports found. Generate a report first."))
		content.WriteString("\n\n")
		content.WriteString(instructionsStyle().Render("Press 'b' to go back"))
	}

	return content.String()
}

// renderGeneratingReport renders the report generation view
func (m Model) renderGeneratingReport() string {
	var content strings.Builder

	spinnerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	content.WriteString(spinnerStyle.Render("🔄 Generating System Health Report..."))
	content.WriteString("\n\n")

	content.WriteString(instructionsStyle().Render("Please wait while we collect system data..."))
	content.WriteString("\n\n")

	// Show what's being collected
	collectingItems := []string{
		"✓ System information",
		"✓ Resource usage (CPU, Memory, Disk)",
		"✓ Security status",
		"✓ Docker container status",
		"✓ Performance metrics",
		"✓ Process information",
	}

	for _, item := range collectingItems {
		content.WriteString(fmt.Sprintf("  %s\n", item))
	}

	return content.String()
}

// renderSavingReport renders the report saving view
func (m Model) renderSavingReport() string {
	var content strings.Builder

	savingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	content.WriteString(savingStyle.Render("💾 Saving Report..."))
	content.WriteString("\n\n")

	content.WriteString(instructionsStyle().Render("Saving the current report to file..."))

	return content.String()
}

// renderViewingReport renders the report viewing view
func (m Model) renderViewingReport() string {
	var content strings.Builder

	// Header with navigation
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	content.WriteString(headerStyle.Render("📄 Viewing Report"))
	content.WriteString("\n\n")

	// Navigation instructions
	navStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Italic(true)

	content.WriteString(navStyle.Render("Use arrow keys to scroll, 'q' to return to report menu, 's' to save"))
	content.WriteString("\n\n")

	// Report content in viewport
	if m.Reporting.CurrentReport != "" {
		// Set viewport content and ensure it's properly sized
		m.Ui.Viewport.SetContent(m.Reporting.CurrentReport)
		m.Ui.Viewport.Height = m.Ui.Height - 8 // Reserve space for header and navigation
		content.WriteString(m.Ui.Viewport.View())
	} else {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("No report content available. Generate a report first."))
	}

	return content.String()
}

// renderSaveNotification renders the save confirmation notification
func (m Model) renderSaveNotification() string {
	var content strings.Builder

	notificationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Padding(0, 1)

	content.WriteString(notificationStyle.Render(m.Reporting.SaveNotification))
	content.WriteString("\n\n")

	return content.String()
}

// instructionsStyle returns a consistent style for instructions
func instructionsStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
}
