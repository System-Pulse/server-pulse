package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderContainerMenu() string {

	if m.Monitor.ContainerMenuState != v.ContainerMenuVisible || m.Monitor.SelectedContainer == nil {
		return ""
	}

	doc := strings.Builder{}

	doc.WriteString("CONTAINER MENU\n")
	doc.WriteString(fmt.Sprintf("Container: %s\n", m.Monitor.SelectedContainer.Name))
	doc.WriteString(fmt.Sprintf("Status: %s\n", m.Monitor.SelectedContainer.Status))
	doc.WriteString("\n")

	for i, item := range m.Monitor.ContainerMenuItems {
		prefix := "  "
		if i == m.Monitor.SelectedMenuItem {
			prefix = "> "
		}
		doc.WriteString(fmt.Sprintf("%s[%s] %s\n", prefix, item.Key, item.Label))
	}

	doc.WriteString("\n")
	doc.WriteString("Navigation: ↑↓ Navigate • Enter Select • ESC Close\n")

	return v.MenuStyle.Render(doc.String())
}

func (m Model) renderContainerSingleView() string {

	if m.Monitor.SelectedContainer == nil {
		return v.CardStyle.Render("No container selected.")
	}
	style := lipgloss.NewStyle().Padding(0, 2).
		Foreground(v.WhiteColor).
		Background(v.SuccessColor).
		Bold(true)
	tabs := renderNav(m.Monitor.ContainerTabs, m.ContainerTab, style)

	tabsHeader := lipgloss.JoinHorizontal(lipgloss.Top, tabs)

	var content string
	switch m.ContainerTab {
	case model.ContainerTabGeneral:
		content = m.renderContainerGeneral()
	case model.ContainerTabCPU:
		content = m.renderContainerCPU()
	case model.ContainerTabMemory:
		content = m.renderContainerMemory()
	case model.ContainerTabNetwork:
		content = m.renderContainerNetwork()
	// case model.ContainerTabDisk:
	// 	content = m.renderContainerDisk() // temporary
	case model.ContainerTabEnv:
		content = m.renderContainerEnv()
	default:
		content = m.renderContainerGeneral()
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabsHeader, content)
}

func (m Model) renderContainerGeneral() string {
	doc := strings.Builder{}

	containerName := "N/A"
	if m.Monitor.ContainerDetails != nil {
		containerName = m.Monitor.ContainerDetails.Name
	} else if m.Monitor.SelectedContainer != nil {
		containerName = m.Monitor.SelectedContainer.Name
	}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(fmt.Sprintf("Container: %s", containerName)))
	doc.WriteString("\n\n")

	if m.Monitor.ContainerDetails != nil {
		info := fmt.Sprintf("ID: %s\nName: %s\nImage: %s\nStatus: %s\nProject: %s\nCreated: %s\nUptime: %s\nHealth: %s\nIP Address: %s\nGateway: %s",
			m.Monitor.ContainerDetails.ID,
			m.Monitor.ContainerDetails.Name,
			m.Monitor.ContainerDetails.Image,
			getStatusWithIcon(m.Monitor.ContainerDetails.Status),
			m.Monitor.ContainerDetails.Project,
			m.Monitor.ContainerDetails.CreatedAt,
			m.Monitor.ContainerDetails.Uptime,
			getHealthWithIcon(m.Monitor.ContainerDetails.HealthCheck),
			m.Monitor.ContainerDetails.IPAddress,
			m.Monitor.ContainerDetails.Gateway,
		)
		doc.WriteString(v.MetricLabelStyle.Render(info))

		if len(m.Monitor.ContainerDetails.Ports) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Ports:"))
			doc.WriteString("\n")
			for _, port := range m.Monitor.ContainerDetails.Ports {
				portInfo := fmt.Sprintf("  %s:%d → %d/%s", port.HostIP, port.PublicPort, port.PrivatePort, port.Type)
				doc.WriteString(v.MetricValueStyle.Render(portInfo))
				doc.WriteString("\n")
			}
		}

		if m.Monitor.ContainerDetails.Command != "" {
			doc.WriteString("\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Command:"))
			doc.WriteString("\n")
			doc.WriteString(v.MetricValueStyle.Render("  " + m.Monitor.ContainerDetails.Command))
		}

		if len(m.Monitor.ContainerDetails.Environment) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Environment:"))
			doc.WriteString(fmt.Sprintf("  %d variables", len(m.Monitor.ContainerDetails.Environment)))
		}
	} else {
		info := "Loading container details..."
		doc.WriteString(v.MetricLabelStyle.Render(info))
	}

	return v.CardStyle.Render(doc.String())
}

func (m Model) renderContainerCPU() string {

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("CPU Usage"))
	doc.WriteString("\n")

	if m.Monitor.ContainerDetails != nil {
		cpuPercent := m.Monitor.ContainerDetails.Stats.CPUPercent
		doc.WriteString(fmt.Sprintf("Usage: %.1f%%\n", cpuPercent))

		// Display progress bar with dynamic color
		doc.WriteString(renderProgressBar(cpuPercent) + "\n\n")

		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Usage History:"))
		doc.WriteString("\n")
		chart := m.renderCPUChart(50, 10)
		doc.WriteString(chart)

		// Display per-core CPU usage if available
		doc.WriteString("\n\n")
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Per-Core Usage:"))
		doc.WriteString("\n")
		perCoreCharts := m.renderAllPerCPUCharts(50, 6)
		doc.WriteString(perCoreCharts)
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("Loading CPU metrics..."))
	}

	return v.CardStyle.Render(doc.String())
}

func (m Model) renderContainerMemory() string {

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Memory Usage"))
	doc.WriteString("\n")

	if m.Monitor.ContainerDetails != nil {
		memPercent := m.Monitor.ContainerDetails.Stats.MemoryPercent
		memUsage := m.Monitor.ContainerDetails.Stats.MemoryUsage
		memLimit := m.Monitor.ContainerDetails.Stats.MemoryLimit

		doc.WriteString(fmt.Sprintf("Usage: %.1f%% | Used: %s | Limit: %s | Available: %s\n\n", memPercent,
			utils.FormatBytes(memUsage), utils.FormatBytes(memLimit), utils.FormatBytes(memLimit-memUsage)))

		// Display progress bar with dynamic color
		doc.WriteString(renderProgressBar(memPercent) + "\n\n")

		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Usage History:"))
		doc.WriteString("\n")
		chart := m.renderMemoryChart(50, 10)
		doc.WriteString(chart)
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("Loading memory metrics..."))
	}

	return v.CardStyle.Render(doc.String())
}

func (m Model) renderContainerNetwork() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Network Usage"))
	doc.WriteString("\n")

	if m.Monitor.ContainerDetails != nil {
		rxBytes := m.Monitor.ContainerDetails.Stats.NetworkRx
		txBytes := m.Monitor.ContainerDetails.Stats.NetworkTx

		// doc.WriteString(fmt.Sprintf("RX: %s/s\n", utils.FormatBytes(rxBytes)))
		// doc.WriteString(fmt.Sprintf("TX: %s/s\n\n", utils.FormatBytes(txBytes)))
		doc.WriteString(fmt.Sprintf("RX Total: %s\n", utils.FormatBytes(rxBytes)))
		doc.WriteString(fmt.Sprintf("TX Total: %s\n\n", utils.FormatBytes(txBytes)))
		// doc.WriteString("\n")
		rxChart := m.renderNetworkRXChart(50, 6)
		doc.WriteString(rxChart)

		doc.WriteString("\n")
		txChart := m.renderNetworkTXChart(50, 6)
		doc.WriteString(txChart)

		if m.Monitor.ContainerDetails.IPAddress != "" {
			doc.WriteString("\n\n" + lipgloss.NewStyle().Bold(true).Render("Network Interfaces:"))
			doc.WriteString("\n")
			doc.WriteString(v.MetricValueStyle.Render("  " + m.Monitor.ContainerDetails.IPAddress))
		}
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("Loading network metrics..."))
	}

	return v.CardStyle.Render(doc.String())
}

/* temporary
func (m Model) renderContainerDisk() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Disk I/O"))
	doc.WriteString("\n\n")

	if m.Monitor.ContainerDetails != nil {
		readBytes := m.Monitor.ContainerDetails.Stats.BlockRead
		writeBytes := m.Monitor.ContainerDetails.Stats.BlockWrite

		doc.WriteString(fmt.Sprintf("Read: %s\n", utils.FormatBytes(readBytes)))
		doc.WriteString(fmt.Sprintf("Write: %s\n\n", utils.FormatBytes(writeBytes)))

		totalIO := readBytes + writeBytes
		if totalIO > 0 {
			readPercent := float64(readBytes) / float64(totalIO) * 100
			writePercent := float64(writeBytes) / float64(totalIO) * 100

			readBlocks := int((float64(readBytes) / float64(totalIO)) * 20)
			writeBlocks := int((float64(writeBytes) / float64(totalIO)) * 20)

			readBar := strings.Repeat("█", readBlocks) + strings.Repeat("░", 20-readBlocks)
			writeBar := strings.Repeat("█", writeBlocks) + strings.Repeat("░", 20-writeBlocks)

			doc.WriteString(fmt.Sprintf("READ  [%s] %.1f%%\n", readBar, readPercent))
			doc.WriteString(fmt.Sprintf("WRITE [%s] %.1f%%\n", writeBar, writePercent))
		}

		doc.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("I/O History:"))
		doc.WriteString("\n")
		doc.WriteString("Disk I/O history charts coming soon...")
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("Loading disk metrics..."))
	}

	return v.CardStyle.Render(doc.String())
}
*/

func (m Model) renderContainerEnv() string {

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Environment Variables"))
	doc.WriteString("\n\n")

	if m.Monitor.ContainerDetails != nil && len(m.Monitor.ContainerDetails.Environment) > 0 {
		for _, env := range m.Monitor.ContainerDetails.Environment {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				if len(value) > 50 {
					value = value[:47] + "..."
				}
				doc.WriteString(fmt.Sprintf("%s: %s\n",
					v.MetricLabelStyle.Render(key),
					v.MetricValueStyle.Render(value)))
			}
		}
	} else if m.Monitor.ContainerDetails != nil {
		doc.WriteString(v.MetricLabelStyle.Render("No environment variables found"))
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("Loading environment variables..."))
	}

	return v.CardStyle.Render(doc.String())
}

func (m Model) renderContainerLogs() string {
	if m.Monitor.SelectedContainer == nil {
		return v.CardStyle.Render("No container selected")
	}

	doc := strings.Builder{}
	status := "Loading..."
	if m.Monitor.ContainerLogsStreaming {
		status = "🟢 Live"
	} else if !m.Monitor.ContainerLogsLoading {
		if strings.ToLower(m.Monitor.SelectedContainer.Status) == "up" {
			status = "🟡 Press 's' for live"
		} else {
			status = "🔴 Streaming unavailable (container not running)"
		}
	} else {
		status = "⏳ Loading..."
	}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(
		fmt.Sprintf("Logs: %s | %s (Page %d/%d)",
			m.Monitor.SelectedContainer.Name,
			status,
			m.Monitor.ContainerLogsPagination.CurrentPage,
			m.Monitor.ContainerLogsPagination.TotalPages)))

	if m.Monitor.ContainerLogsLoading {
		doc.WriteString("\n\n" + v.MetricLabelStyle.Render("Loading logs..."))
	} else if len(m.Monitor.ContainerLogsPagination.Lines) > 0 {
		logStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("235")).
			Height(m.getContentHeight() - 4).
			Padding(1).
			Width(80)

		m.LogsViewport.Style = logStyle
		doc.WriteString("\n" + m.LogsViewport.View())
	} else {
		doc.WriteString("\n\n" + v.MetricLabelStyle.Render("No logs available"))
	}

	return v.CardStyle.Render(doc.String())
}

func getStatusWithIcon(status string) string {
	switch status {
	case "running":
		return "▶ " + status
	case "exited":
		return "⏹ " + status
	case "paused":
		return "⏸ " + status
	case "created":
		return "◉ " + status
	default:
		return status
	}
}

func getHealthWithIcon(health string) string {
	switch health {
	case "healthy":
		return "☼ " + health
	case "unhealthy":
		return "⚠ " + health
	case "starting":
		return "◌ " + health
	default:
		return health
	}
}

// renderProgressBar creates a progress bar with appropriate color based on percentage
func renderProgressBar(percent float64) string {
	var gradientStart, gradientEnd string

	switch {
	case percent < 50:
		gradientStart, gradientEnd = "#00ff00", "#00cc00" // Green
	case percent < 80:
		gradientStart, gradientEnd = "#ffff00", "#cccc00" // Yellow
	default:
		gradientStart, gradientEnd = "#ff0000", "#cc0000" // Red
	}

	// Create a temporary progress bar with the desired gradient
	// Use the same width as other progress bars in the system
	prog := progress.New(progress.WithWidth(v.ProgressBarWidth), progress.WithGradient(gradientStart, gradientEnd))
	return prog.ViewAs(percent / 100)
}
