package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/System-Pulse/server-pulse/widgets/model"
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
	case model.ContainerTabDisk:
		content = m.renderContainerDisk()
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
	doc.WriteString("\n\n")

	if m.Monitor.ContainerDetails != nil {
		cpuPercent := m.Monitor.ContainerDetails.Stats.CPUPercent
		doc.WriteString(fmt.Sprintf("Usage: %.1f%%\n", cpuPercent))

		cpuBar := ""
		cpuBlocks := int(cpuPercent / 5)
		barColor := getCPUColor(cpuPercent)

		for i := range 20 {
			if i < cpuBlocks {
				cpuBar += "█"
			} else {
				cpuBar += "░"
			}
		}

		coloredBar := lipgloss.NewStyle().Foreground(barColor).Render(cpuBar)
		doc.WriteString(fmt.Sprintf("[%s]\n\n", coloredBar))

		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Usage History:"))
		doc.WriteString("\n")
		chart := m.renderCPUChart(50, 10)
		doc.WriteString(chart)
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("Loading CPU metrics..."))
	}

	return v.CardStyle.Render(doc.String())
}

func (m Model) renderContainerMemory() string {

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Memory Usage"))
	doc.WriteString("\n\n")

	if m.Monitor.ContainerDetails != nil {
		memPercent := m.Monitor.ContainerDetails.Stats.MemoryPercent
		memUsage := m.Monitor.ContainerDetails.Stats.MemoryUsage
		memLimit := m.Monitor.ContainerDetails.Stats.MemoryLimit

		doc.WriteString(fmt.Sprintf("Usage: %.1f%%\n", memPercent))
		doc.WriteString(fmt.Sprintf("Used: %s\n", utils.FormatBytes(memUsage)))
		doc.WriteString(fmt.Sprintf("Limit: %s\n", utils.FormatBytes(memLimit)))
		doc.WriteString(fmt.Sprintf("Available: %s\n\n", utils.FormatBytes(memLimit-memUsage)))

		memBar := ""
		memBlocks := int(memPercent / 5)
		barColor := getMemoryColor(memPercent)

		for i := range 20 {
			if i < memBlocks {
				memBar += "█"
			} else {
				memBar += "░"
			}
		}

		coloredBar := lipgloss.NewStyle().Foreground(barColor).Render(memBar)
		doc.WriteString(fmt.Sprintf("[%s]\n\n", coloredBar))

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
	doc.WriteString("\n\n")

	if m.Monitor.ContainerDetails != nil {
		rxBytes := m.Monitor.ContainerDetails.Stats.NetworkRx
		txBytes := m.Monitor.ContainerDetails.Stats.NetworkTx

		doc.WriteString(fmt.Sprintf("RX: %s/s\n", utils.FormatBytes(rxBytes)))
		doc.WriteString(fmt.Sprintf("TX: %s/s\n\n", utils.FormatBytes(txBytes)))

		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Receive Traffic (MB/s):"))
		doc.WriteString("\n")
		rxChart := m.renderNetworkRXChart(50, 6)
		doc.WriteString(rxChart)

		doc.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("Transmit Traffic (MB/s):"))
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

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(fmt.Sprintf("Logs: %s", m.Monitor.SelectedContainer.Name)))
	doc.WriteString("\n\n")

	if m.Monitor.ContainerLogsLoading {
		doc.WriteString(v.MetricLabelStyle.Render("Loading logs..."))
	} else if m.Monitor.ContainerLogs != "" {
		logStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("235")).
			Padding(1).
			Height(20).
			Width(80)

		m.Ui.Viewport.SetContent(m.Monitor.ContainerLogs)
		m.Ui.Viewport.Style = logStyle

		doc.WriteString(m.Ui.Viewport.View())
	} else {
		doc.WriteString(v.MetricLabelStyle.Render("No logs available or logs are empty"))
	}

	doc.WriteString("\n\n")
	doc.WriteString(v.MetricLabelStyle.Render("Press 'r' to refresh logs, 'b' to go back"))

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

func getCPUColor(percent float64) lipgloss.Color {
	switch {
	case percent < 50:
		return lipgloss.Color("46") // Green
	case percent < 80:
		return lipgloss.Color("226") // Yellow
	default:
		return lipgloss.Color("196") // Red
	}
}

func getMemoryColor(percent float64) lipgloss.Color {
	switch {
	case percent < 60:
		return lipgloss.Color("46") // Green
	case percent < 85:
		return lipgloss.Color("226") // Yellow
	default:
		return lipgloss.Color("196") // Red
	}
}
