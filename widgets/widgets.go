package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/utils"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateProcessTable() tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.searchInput.Value())

	for _, p := range m.processes {
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
	m.processTable.SetRows(rows)
	return nil
}

func (m *Model) updateContainerTable(containers []app.Container) tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.searchInput.Value())

	for _, c := range containers {
		if searchTerm != "" && !strings.Contains(strings.ToLower(c.Image), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Name), searchTerm) &&
			!strings.Contains(strings.ToLower(c.ID), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Status), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Project), searchTerm) {
			continue
		}

		// Format status with icon like ctop
		statusWithIcon := m.getStatusWithIconForTable(c.Status, c.Health)

		rows = append(rows, table.Row{
			c.ID,
			utils.Ellipsis(c.Image, 12),
			utils.Ellipsis(c.Name, 16),
			statusWithIcon,
			c.Project,
			utils.Ellipsis(c.PortsStr, 20),
		})
	}
	m.container.SetRows(rows)
	return nil
}

func (m *Model) getStatusWithIconForTable(status, health string) string {
	switch {
	case health == "healthy":
		return "☼ " + status
	case health == "unhealthy":
		return "⚠ " + status
	case status == "running":
		return "▶ " + status
	case status == "exited":
		return "⏹ " + status
	case status == "paused":
		return "⏸ " + status
	case status == "created":
		return "◉ " + status
	default:
		return status
	}
}

// handleRealTimeStats traite les statistiques en temps réel d'un conteneur
func (m *Model) handleRealTimeStats(containerID string, statsChan chan app.ContainerStatsMsg) {
	for stats := range statsChan {
		// Mettre à jour les données en temps réel pour l'affichage
		if m.containerDetails != nil && m.containerDetails.ID == containerID {
			m.containerDetails.Stats.CPUPercent = stats.CPUPercent
			m.containerDetails.Stats.MemoryPercent = stats.MemPercent
			m.containerDetails.Stats.MemoryUsage = stats.MemUsage
			m.containerDetails.Stats.MemoryLimit = stats.MemLimit
			m.containerDetails.Stats.NetworkRx = stats.NetRX
			m.containerDetails.Stats.NetworkTx = stats.NetTX

			// Mettre à jour les graphiques en temps réel
			m.updateChartsWithStats(stats)
		}
	}
}

// updateChartsWithStats met à jour les graphiques avec les nouvelles statistiques
func (m *Model) updateChartsWithStats(stats app.ContainerStatsMsg) {
	now := time.Now()

	// Mettre à jour l'historique CPU
	m.cpuHistory.Points = append(m.cpuHistory.Points, DataPoint{
		Timestamp: now,
		Value:     stats.CPUPercent,
	})
	if len(m.cpuHistory.Points) > m.cpuHistory.MaxPoints {
		m.cpuHistory.Points = m.cpuHistory.Points[1:]
	}

	// Mettre à jour l'historique mémoire
	m.memoryHistory.Points = append(m.memoryHistory.Points, DataPoint{
		Timestamp: now,
		Value:     stats.MemPercent,
	})
	if len(m.memoryHistory.Points) > m.memoryHistory.MaxPoints {
		m.memoryHistory.Points = m.memoryHistory.Points[1:]
	}

	// Mettre à jour l'historique réseau (convertir bytes/s en MB/s)
	m.networkRxHistory.Points = append(m.networkRxHistory.Points, DataPoint{
		Timestamp: now,
		Value:     float64(stats.NetRX) / 1024 / 1024,
	})
	if len(m.networkRxHistory.Points) > m.networkRxHistory.MaxPoints {
		m.networkRxHistory.Points = m.networkRxHistory.Points[1:]
	}

	m.networkTxHistory.Points = append(m.networkTxHistory.Points, DataPoint{
		Timestamp: now,
		Value:     float64(stats.NetTX) / 1024 / 1024,
	})
	if len(m.networkTxHistory.Points) > m.networkTxHistory.MaxPoints {
		m.networkTxHistory.Points = m.networkTxHistory.Points[1:]
	}

	m.lastChartUpdate = now
}

func (m *Model) loadContainerDetails(containerID string) tea.Cmd {
	return func() tea.Msg {
		details, err := m.app.GetContainerDetails(containerID)
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


// Commande pour effacer le message d'opération après un délai
func clearOperationMessage() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return ClearOperationMsg{}
	})
}
