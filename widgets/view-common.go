package widgets

import (
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/utils"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func getHealthColor(health string) lipgloss.Color {
	switch health {
	case "healthy", "running":
		return lipgloss.Color("10")
	case "unhealthy":
		return lipgloss.Color("9")
	case "starting":
		return lipgloss.Color("11")
	case "exited":
		return lipgloss.Color("1")
	default:
		return lipgloss.Color("7")
	}
}

func getHealthIcon(health string) string {
	switch health {
	case "healthy":
		return "☼ "
	case "unhealthy":
		return "⚠ "
	case "starting":
		return "◌ "
	case "running":
		return "▶ "
	case "exited":
		return "⏹ "
	case "paused":
		return "⏸ "
	case "created":
		return "◉ "
	default:
		return ""
	}
}

func (m *Model) getStatusWithIconForTable(status, health string) (string, string) {
	icon := getHealthIcon(health)
	displayHealth := health

	switch health {
	case "running", "exited", "paused", "created":
		displayHealth = "N/A"
	}

	color := getHealthColor(health)
	style := lipgloss.NewStyle().Foreground(color)

	var displayText string
	if icon != "" {
		displayText = style.Render(icon) + status
	} else {
		displayText = style.Render(status)
	}

	return displayText, displayHealth
}

func clearOperationMessage() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return model.ClearOperationMsg{}
	})
}

func (m *Model) loadContainerDetails(containerID string) tea.Cmd {
	return func() tea.Msg {
		details, err := m.Monitor.App.GetContainerDetails(containerID)
		if err != nil {
			return utils.ErrMsg(err)
		}
		return app.ContainerDetailsMsg(*details)
	}
}

/*
"1" - Red
"2" - Green
"3" - Yellow
"4" - Blue
"5" - Magenta
"6" - Cyan
"7" - Gray/White
"8" - Dark Gray
"9" - Light Red
"10" - Light Green
"11" - Light Yellow
"12" - Light Blue
"13" - Light Magenta
"14" - Light Cyan
"15" - White
*/
