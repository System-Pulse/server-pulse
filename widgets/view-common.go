package widgets

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/system/logs"
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

	// Initialize container history if it doesn't exist
	if _, exists := m.Monitor.ContainerHistories[stats.ContainerID]; !exists {
		m.Monitor.ContainerHistories[stats.ContainerID] = model.ContainerHistory{
			CpuHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			PerCpuHistory: make(map[int]model.DataHistory),
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

	// Update CPU history
	containerHistory.CpuHistory.Points = append(containerHistory.CpuHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     stats.CPUPercent,
	})
	if len(containerHistory.CpuHistory.Points) > containerHistory.CpuHistory.MaxPoints {
		containerHistory.CpuHistory.Points = containerHistory.CpuHistory.Points[1:]
	}

	// Update per-core CPU history
	if len(stats.PerCPUPercents) > 0 {
		for i, cpuPercent := range stats.PerCPUPercents {
			// Initialize history for this core if it doesn't exist
			if _, exists := containerHistory.PerCpuHistory[i]; !exists {
				containerHistory.PerCpuHistory[i] = model.DataHistory{
					MaxPoints: 60,
					Points:    make([]model.DataPoint, 0),
				}
			}

			coreHistory := containerHistory.PerCpuHistory[i]
			coreHistory.Points = append(coreHistory.Points, model.DataPoint{
				Timestamp: now,
				Value:     cpuPercent,
			})
			if len(coreHistory.Points) > coreHistory.MaxPoints {
				coreHistory.Points = coreHistory.Points[1:]
			}
			containerHistory.PerCpuHistory[i] = coreHistory
		}
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

func (m *Model) updateFirewallTable() tea.Cmd {
	var rows []table.Row

	for _, rule := range m.Diagnostic.FirewallInfo.Rules {
		rows = append(rows, table.Row{
			rule.Description,
		})
	}

	m.Diagnostic.FirewallTable.SetRows(rows)
	return nil
}

func (m *Model) updateAutoBanTable() tea.Cmd {
	var rows []table.Row

	for _, jail := range m.Diagnostic.AutoBanInfo.Jails {
		description := fmt.Sprintf("[%s] %s | Filter: %s | Currently: %d | Total: %d | %s",
			jail.Name, jail.Status, jail.Filter, jail.CurrentBans, jail.TotalBans, jail.Details)
		rows = append(rows, table.Row{
			description,
		})
	}

	m.Diagnostic.AutoBanTable.SetRows(rows)
	return nil
}

func (m *Model) updatePortsTable() tea.Cmd {
	var rows []table.Row

	for _, port := range m.Diagnostic.OpenedPortsInfo.Ports {
		cmd := exec.Command("sudo", "-n", "lsof", "-i", fmt.Sprintf(":%d", port))
		output, err := cmd.Output()
		if err != nil {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", port),
				"Unknown",
				"TCP",
				"",
			})
			continue
		}

		service := "Unknown"
		protocol := "TCP"
		pid := ""

		lines := strings.SplitSeq(string(output), "\n")
		for line := range lines {
			if strings.Contains(line, "LISTEN") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					service = fields[0]
					pid = fields[1]

					if len(fields) > 8 {
						nameField := fields[8]
						if strings.Contains(nameField, "TCP") {
							protocol = "TCP"
						} else if strings.Contains(nameField, "UDP") {
							protocol = "UDP"
						}
					}
				}
				break
			}
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", port),
			service,
			protocol,
			pid,
		})
	}

	m.Diagnostic.PortsTable.SetRows(rows)
	return nil
}

// updateLogsTable updates the logs table with current log entries
func (m Model) updateLogsTable() tea.Cmd {
	if m.Diagnostic.LogsInfo == nil {
		return nil
	}

	rows := []table.Row{}
	for _, entry := range m.Diagnostic.LogsInfo.Entries {
		// Format timestamp
		timestamp := "N/A"
		if !entry.Timestamp.IsZero() {
			timestamp = entry.Timestamp.Format("2006-01-02 15:04:05")
		}

		// Truncate message if too long
		message := entry.Message
		if len(message) > 70 {
			message = message[:67] + "..."
		}

		// Default empty values if missing
		level := entry.Level
		if level == "" {
			level = "INFO"
		}

		service := entry.Service
		if service == "" {
			service = "-"
		}

		rows = append(rows, table.Row{
			timestamp,
			level,
			service,
			message,
		})
	}

	m.Diagnostic.LogsTable.SetRows(rows)
	return nil
}

// loadLogs loads system logs with current filters
func (m Model) loadLogs() tea.Cmd {
	// Update LogManager with current auth state
	m.Diagnostic.LogManager.CanUseSudo = m.Diagnostic.CanRunSudo
	m.Diagnostic.LogManager.SudoPassword = m.Diagnostic.SecurityManager.SudoPassword

	return m.Diagnostic.LogManager.GetSystemLogs(m.Diagnostic.LogFilters)
}

// applyTimeRangeSelection updates the time range filter based on selection
func (m *Model) applyTimeRangeSelection() {
	timeRanges := []string{"", "1h", "24h", "7d", ""}
	if m.Diagnostic.LogTimeSelected < len(timeRanges) {
		m.Diagnostic.LogFilters.TimeRange = timeRanges[m.Diagnostic.LogTimeSelected]
	}
}

// applyLogLevelSelection updates the log level filter based on selection
func (m *Model) applyLogLevelSelection() {
	levels := []logs.LogLevel{
		logs.LogLevelAll,
		logs.LogLevelError,
		logs.LogLevelWarning,
		logs.LogLevelInfo,
		logs.LogLevelDebug,
	}
	if m.Diagnostic.LogLevelSelected < len(levels) {
		m.Diagnostic.LogFilters.Level = levels[m.Diagnostic.LogLevelSelected]
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
