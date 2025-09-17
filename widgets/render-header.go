package widgets

import (
	"fmt"
	v "github.com/System-Pulse/server-pulse/widgets/vars"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHeader() string {
	var menu []string

	header := []string{}
	headerStyle := lipgloss.NewStyle().
		MarginLeft(5).
		Padding(0, 1).
		Foreground(lipgloss.Color("255")).
		Background(v.SuccessColor).
		Bold(true)
	header = append(header, headerStyle.Render(v.AsciiArt))

	for i, t := range m.Ui.Tabs.DashBoard {
		style := lipgloss.NewStyle()
		if i == m.Ui.SelectedTab {
			if m.Ui.ActiveView == i {
				style = v.CardButtonStyle.
					Foreground(lipgloss.Color("229")).
					Background(v.SuccessColor)
			} else {
				style = v.CardButtonStyle.
					Foreground(lipgloss.Color("229"))
			}
		} else {
			style = v.CardButtonStyleDesactive.
				Foreground(lipgloss.Color("240"))
		}
		menu = append(menu, style.Render(t))
	}

	doc := strings.Builder{}
	systemInfo := fmt.Sprintf("Host: %s	|	OS: %s	|	Kernel: %s	|	Uptime: %s", m.Monitor.System.Hostname, m.Monitor.System.OS, m.Monitor.System.Kernel, utils.FormatUptime(m.Monitor.System.Uptime))
	doc.WriteString(v.MetricLabelStyle.Render(systemInfo))
	header = append(header, v.CardStyle.MarginBottom(0).Render(doc.String()))

	header = append(header, lipgloss.JoinHorizontal(lipgloss.Top, menu...))
	return lipgloss.JoinVertical(lipgloss.Top, header...)
}
