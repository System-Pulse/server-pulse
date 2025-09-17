package widgets

import (
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/utils"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/table"
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

func (m *Model) updateChartsWithStats(stats app.ContainerStatsMsg) {
	now := time.Now()

	m.Monitor.CpuHistory.Points = append(m.Monitor.CpuHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     stats.CPUPercent,
	})
	if len(m.Monitor.CpuHistory.Points) > m.Monitor.CpuHistory.MaxPoints {
		m.Monitor.CpuHistory.Points = m.Monitor.CpuHistory.Points[1:]
	}

	m.Monitor.MemoryHistory.Points = append(m.Monitor.MemoryHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     stats.MemPercent,
	})
	if len(m.Monitor.MemoryHistory.Points) > m.Monitor.MemoryHistory.MaxPoints {
		m.Monitor.MemoryHistory.Points = m.Monitor.MemoryHistory.Points[1:]
	}

	m.Monitor.NetworkRxHistory.Points = append(m.Monitor.NetworkRxHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     float64(stats.NetRX) / 1024 / 1024,
	})
	if len(m.Monitor.NetworkRxHistory.Points) > m.Monitor.NetworkRxHistory.MaxPoints {
		m.Monitor.NetworkRxHistory.Points = m.Monitor.NetworkRxHistory.Points[1:]
	}

	m.Monitor.NetworkTxHistory.Points = append(m.Monitor.NetworkTxHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     float64(stats.NetTX) / 1024 / 1024,
	})
	if len(m.Monitor.NetworkTxHistory.Points) > m.Monitor.NetworkTxHistory.MaxPoints {
		m.Monitor.NetworkTxHistory.Points = m.Monitor.NetworkTxHistory.Points[1:]
	}

	m.LastChartUpdate = now
}

func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return utils.TickMsg(t)
	})
}

func (m *Model) updateSecurityTable() tea.Cmd {
	var rows []table.Row

	for _, check := range m.Diagnostic.SecurityChecks {
		// Add status icons based on status
		statusWithIcon := m.getSecurityStatusIcon(check.Status) + " " + check.Status

		rows = append(rows, table.Row{
			check.Name,
			statusWithIcon,
			check.Details,
		})
	}

	m.Diagnostic.SecurityTable.SetRows(rows)
	return nil
}

func (m *Model) getSecurityStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "valid", "secure", "disabled", "ok":
		return "✓" // Green checkmark
	case "warning", "expiring":
		return "⚠" // Warning triangle
	case "invalid", "critical", "enabled", "error":
		return "✗" // Red X
	default:
		return "●" // Neutral dot
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
