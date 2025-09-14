package widgets

import (
	"strings"

	"github.com/System-Pulse/server-pulse/widgets/model"
	v "github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/lipgloss"
)

func renderNav(header []string, state model.ContainerTab, styleColor lipgloss.Style) string {
	var tabs []string
	for i, tab := range header {
		style := lipgloss.NewStyle().Padding(0, 2)
		if model.ContainerTab(i) == state {
			style = styleColor
		} else {
			style = style.
				Foreground(lipgloss.Color("240")).
				Background(lipgloss.Color("236"))
		}
		tabs = append(tabs, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderCurrentNav() string {
	if strings.HasPrefix(string(m.Ui.State), "monitor") {
		style := lipgloss.NewStyle().Padding(0, 2).
			Foreground(v.ClearWhite).
			Background(v.PurpleCollor).
			Bold(true)
		return renderNav(m.Ui.Tabs.Monitor, model.ContainerTab(m.Ui.SelectedMonitor), style)
	}

	if m.Ui.State == model.StateNetwork {
		style := lipgloss.NewStyle().Padding(0, 2).
			Foreground(v.ClearWhite).
			Background(v.PurpleCollor).
			Bold(true)
		return renderNav(m.Network.Nav, model.ContainerTab(m.Network.SelectedItem), style)
	}
	if m.Ui.State == model.StateDiagnostics {
		style := lipgloss.NewStyle().Padding(0, 2).
			Foreground(v.ClearWhite).
			Background(v.PurpleCollor).
			Bold(true)
		return renderNav(m.Diagnostic.Nav, model.ContainerTab(m.Diagnostic.SelectedItem), style)
	}
	return ""
}
