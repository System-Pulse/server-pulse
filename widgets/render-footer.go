package widgets

import (
	"fmt"
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
	switch m.Ui.State {
	case model.StateHome:
		hints = "[Enter] Select • [Tab/←→] Navigate • [1-4] Quick select • [q] Quit"
	case model.StateMonitor:
		hints = "[Enter] Select • [Tab/←→] Navigate • [1-3] Quick select • [b] Back • [q] Quit"
	case model.StateSystem:
		hints = "[↑↓] Scroll • [b] Back • [q] Quit"
	case model.StateProcess:
		hints = "[↑↓] Navigate • [/] Search • [k] Kill • [s/m] Sort • [b] Back • [q] Quit"
	case model.StateContainers:
		hints = "[↑↓] Navigate • [Enter] Menu • [/] Search • [b] Back • [q] Quit"
	case model.StateContainer:
		hints = "[Tab/←→] Switch tabs • [b] Back • [q] Quit"
	case model.StateContainerLogs:
		streamingHint := ""
		if m.Monitor.SelectedContainer != nil {
			if strings.ToLower(m.Monitor.SelectedContainer.Status) == "up" {
				if m.Monitor.ContainerLogsStreaming {
					streamingHint = " | [s] Stop streaming"
				} else {
					streamingHint = " | [s] Start streaming"
				}
			} else {
				streamingHint = ""
			}
		}
		hints = fmt.Sprintf("[↑↓] Scroll • [r] Refresh • [home/end] Navigate%s • [b] Back • [q] Quit", streamingHint)
	case model.StateNetwork:
		hints = "[Tab/←→] Switch tabs • [b] Back • [q] Quit"
	case model.StateDiagnostics:
		// Check if we're on the logs tab
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			// Check if any input is focused
			if m.Diagnostic.LogSearchInput.Focused() {
				hints = "[ESC] Cancel • [Enter] Apply search • [Type] Search text"
			} else if m.Diagnostic.LogServiceInput.Focused() {
				hints = "[ESC] Cancel • [Enter] Apply filter • [Type] Service name"
			} else if m.Diagnostic.LogTimeRangeInput.Focused() {
				hints = "[ESC] Cancel • [Enter] Apply • [Type] Time range (e.g., '2 hours ago')"
			} else {
				hints = "[←→] Time range • [Shift+←→] Level • [Enter/r] Reload • [/] Search • [s] Service • [↑↓] Navigate • [b] Back"
			}
		} else {
			hints = "[b] Back • [enter] Details • [q] Quit"
		}
	case model.StateCertificateDetails, model.StateSSHRootDetails, model.StateReporting:
		hints = "[b] Back • [q] Quit"
	}

	if m.ConfirmationVisible {
		hints = "[y] Confirm • [n/ESC] Cancel"
	} else if m.Monitor.ContainerMenuState == v.ContainerMenuVisible {
		hints = "[↑↓] Navigate • [Enter] Select • [ESC/b] Close • [o/l/...] Actions"
	}

	return statusLine + "\n" + hints
}
