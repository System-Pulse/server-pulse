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
	return tea.Sequence(
		func() tea.Msg {
			details, err := m.Monitor.App.GetContainerDetails(containerID)
			if err != nil {
				return utils.ErrMsg(err)
			}
			return app.ContainerDetailsMsg(*details)
		},
		func() tea.Msg {
			statsChan, err := m.Monitor.App.GetContainerStatsStream(containerID)
			if err != nil {
				return utils.ErrMsg(err)
			}
			return app.ContainerStatsChanMsg{
				ContainerID: containerID,
				StatsChan:   statsChan,
			}
		},
	)
}

func (m *Model) startContainerStats(containerID string) tea.Cmd {
	return func() tea.Msg {
		statsChan, err := m.Monitor.App.GetContainerStatsStream(containerID)
		if err != nil {
			return utils.ErrMsg(err)
		}

		return app.ContainerStatsChanMsg{
			ContainerID: containerID,
			StatsChan:   statsChan,
		}
	}
}

func (m *Model) stopContainerStats() {
	// Close any active stats channels
	for containerID, containerHistory := range m.Monitor.ContainerHistories {
		// For now, we'll just clear the history when stopping stats
		// In a real implementation, we might want to close the channel
		containerHistory.CpuHistory.Points = []model.DataPoint{}
		containerHistory.MemoryHistory.Points = []model.DataPoint{}
		containerHistory.NetworkRxHistory.Points = []model.DataPoint{}
		containerHistory.NetworkTxHistory.Points = []model.DataPoint{}
		m.Monitor.ContainerHistories[containerID] = containerHistory
	}
}

func (m *Model) updateChartsWithStats(stats app.ContainerStatsMsg) {
	now := time.Now()
	// fmt.Printf("DEBUG: updateChartsWithStats called for container %s - CPU: %.1f%%, Mem: %.1f%%, NetRX: %d, NetTX: %d\n",
	// 	stats.ContainerID, stats.CPUPercent, stats.MemPercent, stats.NetRX, stats.NetTX)

	// Initialize container history if it doesn't exist
	if _, exists := m.Monitor.ContainerHistories[stats.ContainerID]; !exists {
		m.Monitor.ContainerHistories[stats.ContainerID] = model.ContainerHistory{
			CpuHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			MemoryHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			NetworkRxHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			NetworkTxHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
		}
	}

	// Get container history
	containerHistory := m.Monitor.ContainerHistories[stats.ContainerID]
	// fmt.Printf("DEBUG: Container history for %s - CPU points: %d, Memory points: %d\n",
	// 	stats.ContainerID, len(containerHistory.CpuHistory.Points), len(containerHistory.MemoryHistory.Points))

	// Update CPU history
	containerHistory.CpuHistory.Points = append(containerHistory.CpuHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     stats.CPUPercent,
	})
	if len(containerHistory.CpuHistory.Points) > containerHistory.CpuHistory.MaxPoints {
		containerHistory.CpuHistory.Points = containerHistory.CpuHistory.Points[1:]
	}

	// Update Memory history
	containerHistory.MemoryHistory.Points = append(containerHistory.MemoryHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     stats.MemPercent,
	})
	if len(containerHistory.MemoryHistory.Points) > containerHistory.MemoryHistory.MaxPoints {
		containerHistory.MemoryHistory.Points = containerHistory.MemoryHistory.Points[1:]
	}

	// Update Network RX history
	containerHistory.NetworkRxHistory.Points = append(containerHistory.NetworkRxHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     float64(stats.NetRX) / 1024 / 1024,
	})
	if len(containerHistory.NetworkRxHistory.Points) > containerHistory.NetworkRxHistory.MaxPoints {
		containerHistory.NetworkRxHistory.Points = containerHistory.NetworkRxHistory.Points[1:]
	}

	// Update Network TX history
	containerHistory.NetworkTxHistory.Points = append(containerHistory.NetworkTxHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     float64(stats.NetTX) / 1024 / 1024,
	})
	if len(containerHistory.NetworkTxHistory.Points) > containerHistory.NetworkTxHistory.MaxPoints {
		containerHistory.NetworkTxHistory.Points = containerHistory.NetworkTxHistory.Points[1:]
	}

	// Store updated history back to map
	m.Monitor.ContainerHistories[stats.ContainerID] = containerHistory
	// fmt.Printf("DEBUG: Updated container history for %s - CPU points: %d, Memory points: %d\n",
	// 	stats.ContainerID, len(containerHistory.CpuHistory.Points), len(containerHistory.MemoryHistory.Points))
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
