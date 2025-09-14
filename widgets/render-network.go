package widgets

import (
	"fmt"
	"strings"

	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) interfaces() string {
	content := strings.Builder{}

	statusIcon := "🔴"
	statusText := "Disconnected"
	statusColor := v.ErrorColor

	if m.Network.NetworkResource.Connected {
		statusIcon = "🟢"
		statusText = "Connected"
		statusColor = v.SuccessColor
	}

	statusLine := fmt.Sprintf("%s %s",
		statusIcon,
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText))
	content.WriteString(statusLine)

	if len(m.Network.NetworkResource.Interfaces) > 0 {
		tableContent := m.renderTable(m.Network.NetworkTable, "No network interfaces")
		content.WriteString(tableContent)
	}

	return v.CardNetworkStyle.
		Margin(0, 0, 0, 0).
		Padding(0, 1).
		Render(content.String())
}

func (m Model) renderNetwork() string {
	currentView := ""

	switch m.Network.SelectedItem {
	case model.NetworkTabInterface:
		currentView = m.interfaces()
	case model.NetworkTabConnectivity:
		currentView = renderNotImplemented("Connectivity Analysis")
	case model.NetworkTabConfiguration:
		currentView = renderNotImplemented("Network Configuration")
	case model.NetworkTabProtocol:
		currentView = renderNotImplemented("Protocol Analysis")
	}
	return currentView
}
