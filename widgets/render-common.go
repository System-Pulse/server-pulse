package widgets

import (
	"fmt"
	v "github.com/System-Pulse/server-pulse/widgets/vars"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderNotImplemented(feature string) string {
	return v.CardStyle.Render(fmt.Sprintf("🚧 %s\n\nThis feature is not yet implemented.\n\nCheck back in future updates!", feature))
}

func (m Model) renderConfirmationDialog() string {
	if !m.ConfirmationVisible {
		return ""
	}

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Foreground(v.ErrorColor).Render("⚠️  CONFIRMATION REQUIRED"))
	doc.WriteString("\n\n")

	doc.WriteString(v.MetricLabelStyle.Render(m.ConfirmationMessage))
	doc.WriteString("\n\n")

	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Are you sure?"))
	doc.WriteString("\n")
	doc.WriteString("Press 'y' to confirm or 'n' to cancel")

	confirmationStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.ErrorColor).
		Padding(2).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255"))

	return confirmationStyle.Render(doc.String())
}

func (m Model) getAvailableHeight() int {
	if m.Ui.Height <= 0 {
		return 1
	}

	headerHeight := lipgloss.Height(m.renderHeader())
	navHeight := lipgloss.Height(m.renderCurrentNav())
	footerHeight := lipgloss.Height(m.renderFooter())

	availableHeight := m.Ui.Height - headerHeight - navHeight - footerHeight
	return max(1, availableHeight)
}

func (m Model) getContentHeight() int {
	availableHeight := m.getAvailableHeight()

	contentHeight := availableHeight - 2
	return max(1, contentHeight)
}

func (m *Model) cleanupLogsStream() {
	if m.Monitor.LogsCancelFunc != nil {
		m.Monitor.LogsCancelFunc()
	}
	m.Monitor.ContainerLogsStreaming = false
	m.Monitor.ContainerLogsChan = nil
	m.Monitor.LogsCancelFunc = nil
}
