package widgets

import (
	"strings"

	"github.com/System-Pulse/server-pulse/widgets/model"
	v "github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderFooter() string {
	statusLine := ""
	if m.OperationInProgress {
		statusStyle := lipgloss.NewStyle().
			Foreground(v.ClearWhite).
			Background(v.PurpleCollor).
			Padding(0, 1).
			Bold(true)
		statusLine += statusStyle.Render("⏳ Operation in progress...") + "\n"
	} else if m.LastOperationMsg != "" {
		var statusStyle lipgloss.Style
		if strings.Contains(m.LastOperationMsg, "failed") || strings.Contains(m.LastOperationMsg, "Error") {
			statusStyle = lipgloss.NewStyle().
				Foreground(v.WhiteColor).
				Background(v.ErrorColor).
				Padding(0, 1).
				Bold(true)
		} else {
			statusStyle = lipgloss.NewStyle().
				Foreground(v.WhiteColor).
				Background(v.SuccessColor).
				Padding(0, 1).
				Bold(true)
		}
		statusLine += statusStyle.Render(m.LastOperationMsg) + "\n"
	}

	var hints string
	// Use help system for hints, except in special cases
	if m.ConfirmationVisible {
		hints = m.HelpSystem.ViewWithModel(model.StateHome, model.ContainerTab(0), m) // Use base keymap for confirmation
	} else if m.Monitor.ContainerMenuState == v.ContainerMenuVisible {
		hints = m.HelpSystem.ViewWithModel(model.StateHome, model.ContainerTab(0), m) // Use base keymap for container menu
	} else {
		// For diagnostics state, pass the selected diagnostic tab
		if m.Ui.State == model.StateDiagnostics || m.Ui.State == model.StateCertificateDetails || m.Ui.State == model.StateSSHRootDetails {
			hints = m.HelpSystem.ViewWithModel(m.Ui.State, m.Diagnostic.SelectedItem, m)
		} else {
			hints = m.HelpSystem.ViewWithModel(m.Ui.State, model.ContainerTab(0), m)
		}
	}

	return statusLine + "\n" + hints
}
