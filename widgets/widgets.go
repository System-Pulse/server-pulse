package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/utils"
	model "github.com/System-Pulse/server-pulse/widgets/model"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateProcessTable() tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.Ui.SearchInput.Value())

	for _, p := range m.Monitor.Processes {
		if searchTerm != "" && !strings.Contains(strings.ToLower(p.Command), searchTerm) &&
			!strings.Contains(strings.ToLower(p.User), searchTerm) &&
			!strings.Contains(fmt.Sprintf("%d", p.PID), searchTerm) {
			continue
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.PID),
			p.User,
			fmt.Sprintf("%.1f", p.CPU),
			fmt.Sprintf("%.1f", p.Mem),
			utils.Ellipsis(p.Command, 30),
		})
	}
	m.Monitor.ProcessTable.SetRows(rows)
	return nil
}

func (m *Model) updateNetworkTable() tea.Cmd {
	var rows []table.Row

	for _, iface := range m.Network.NetworkResource.Interfaces {
		statusText := "DOWN"
		if iface.Status == "up" {
			statusText = "UP"
		}

		ips := strings.Join(iface.IPs, ", ")
		if ips == "" {
			ips = "No IP"
		}

		rows = append(rows, table.Row{
			iface.Name,
			statusText,
			ips,
			utils.FormatBytes(iface.RxBytes),
			utils.FormatBytes(iface.TxBytes),
		})
	}

	m.Network.NetworkTable.SetRows(rows)

	tableHeight := min(10, len(rows)+1)
	m.Network.NetworkTable.SetHeight(tableHeight)

	return nil
}

func (m *Model) updateContainerTable(containers []app.Container) tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.Ui.SearchInput.Value())

	for _, c := range containers {
		if searchTerm != "" && !strings.Contains(strings.ToLower(c.Image), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Name), searchTerm) &&
			!strings.Contains(strings.ToLower(c.ID), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Status), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Project), searchTerm) {
			continue
		}

		// Format status with icon like ctop
		statusWithIcon, health := m.getStatusWithIconForTable(c.Status, c.Health)

		rows = append(rows, table.Row{
			c.ID,
			utils.Ellipsis(c.Image, 12),
			utils.Ellipsis(c.Name, 16),
			statusWithIcon,
			health,
			c.Project,
			utils.Ellipsis(c.PortsStr, 20),
		})
	}
	m.Monitor.Container.SetRows(rows)
	return nil
}

func (m *Model) getStatusWithIconForTable(status, health string) (string, string) {
	switch health {
	case "healthy":
		return "☼ " + status, health
	case "unhealthy":
		return "⚠ " + status, health
	case "running":
		return "▶ " + status, "N/A"
	case "exited":
		return "⏹ " + status, "N/A"
	case "paused":
		return "⏸ " + status, "N/A"
	case "created":
		return "◉ " + status, "N/A"
	default:
		return status, health
	}
}

func (m *Model) handleRealTimeStats(containerID string, statsChan chan app.ContainerStatsMsg) {
	for stats := range statsChan {
		if m.Monitor.ContainerDetails != nil && m.Monitor.ContainerDetails.ID == containerID {
			m.Monitor.ContainerDetails.Stats.CPUPercent = stats.CPUPercent
			m.Monitor.ContainerDetails.Stats.MemoryPercent = stats.MemPercent
			m.Monitor.ContainerDetails.Stats.MemoryUsage = stats.MemUsage
			m.Monitor.ContainerDetails.Stats.MemoryLimit = stats.MemLimit
			m.Monitor.ContainerDetails.Stats.NetworkRx = stats.NetRX
			m.Monitor.ContainerDetails.Stats.NetworkTx = stats.NetTX

			m.updateChartsWithStats(stats)
		}
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

func (m *Model) loadContainerDetails(containerID string) tea.Cmd {
	return func() tea.Msg {
		details, err := m.Monitor.App.GetContainerDetails(containerID)
		if err != nil {
			return utils.ErrMsg(err)
		}
		return app.ContainerDetailsMsg(*details)
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return utils.TickMsg(t)
	})
}

func clearOperationMessage() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return model.ClearOperationMsg{}
	})
}
