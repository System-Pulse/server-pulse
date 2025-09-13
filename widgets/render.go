package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHeader() string {
	var menu []string
	header := []string{}
	headerStyle := lipgloss.NewStyle().
		MarginLeft(5).
		Padding(0, 1).
		Foreground(lipgloss.Color("255")).
		Background(successColor).
		Bold(true)
	header = append(header, headerStyle.Render(asciiArt))

	for i, t := range m.Ui.Tabs.DashBoard {
		style := lipgloss.NewStyle()
		if i == m.Ui.SelectedTab {
			if m.Ui.ActiveView == i {
				style = cardButtonStyle.
					Foreground(lipgloss.Color("229")).
					Background(successColor)
			} else {
				style = cardButtonStyle.
					Foreground(lipgloss.Color("229"))
			}
		} else {
			style = cardButtonStyleDesactive.
				Foreground(lipgloss.Color("240"))
		}
		menu = append(menu, style.Render(t))
	}

	doc := strings.Builder{}
	systemInfo := fmt.Sprintf("Host: %s	|	OS: %s	|	Kernel: %s	|	Uptime: %s", m.Monitor.System.Hostname, m.Monitor.System.OS, m.Monitor.System.Kernel, utils.FormatUptime(m.Monitor.System.Uptime))
	doc.WriteString(metricLabelStyle.Render(systemInfo))
	header = append(header, cardStyle.MarginBottom(0).Render(doc.String()))

	header = append(header, lipgloss.JoinHorizontal(lipgloss.Top, menu...))
	return lipgloss.JoinVertical(lipgloss.Top, header...)
}

func (m Model) renderNav(header []string, state model.ContainerTab, styleColor lipgloss.Style) string {
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
	if strings.HasPrefix(string(m.Ui.State), "monitor")  {
		style := lipgloss.NewStyle().Padding(0, 2).
			Foreground(clearWhite).
			Background(purpleCollor).
			Bold(true)
		return m.renderNav(m.Ui.Tabs.Monitor, model.ContainerTab(m.Ui.SelectedMonitor), style)
	}

	if m.Ui.State == model.StateNetwork {
		style := lipgloss.NewStyle().Padding(0, 2).
			Foreground(clearWhite).
			Background(purpleCollor).
			Bold(true)
		return m.renderNav(m.Network.Nav, model.ContainerTab(m.Network.SelectedItem), style)
	}
	return ""
}

func (m Model) renderMainContent() string {
	var currentView string

	switch m.Ui.State {
	case model.StateHome:
		currentView = m.renderSystem()
	case model.StateMonitor:
		currentView = m.renderMonitor()
	case model.StateSystem:
		currentView = m.renderSystem()
	case model.StateProcess:
		currentView = m.renderProcesses()
	case model.StateContainers:
		currentView = m.renderContainers()
	case model.StateContainer:
		currentView = m.renderContainerSingleView()
	case model.StateContainerLogs:
		currentView = m.renderContainerLogs()
	case model.StateNetwork:
		currentView = m.renderNetwork()
	case model.StateDiagnostics:
		currentView = m.renderDignostics()
	case model.StateReporting:
		currentView = m.renderReporting()
	default:
		currentView = fmt.Sprintf("Unknown state: %v", m.Ui.State)
	}

	// Utilise le viewport pour le contenu scrollable
	// switch m.Ui.State {
	// case model.StateSystem, model.StateContainerLogs, model.StateHome, model.StateMonitor:
	m.Ui.Viewport.SetContent(currentView)
	return m.Ui.Viewport.View()
	// }

	// return currentView
}

func (m Model) renderMonitor() string {
	var currentView string
	switch m.Ui.SelectedMonitor {
	case 0:
		currentView = m.renderSystem()
	case 1:
		currentView = m.renderProcesses()
	case 2:
		currentView = m.renderContainers()
	}
	return currentView
}

func (m Model) renderContainers() string {
	p := "Search a container..."
	return m.renderTable(m.Monitor.Container, p)
}

func (m Model) renderTable(table table.Model, placeholder string) string {
	content := strings.Builder{}
	m.Ui.SearchInput.Placeholder = placeholder

	if m.Ui.SearchMode {
		searchBar := searchBarStyle.
			Render(m.Ui.SearchInput.View())
		content.WriteString(searchBar)
		content.WriteString("\n")
	}

	content.WriteString(table.View())

	return cardTableStyle.Render(content.String())
}

func (m Model) renderDignostics() string {
	return m.renderNotImplemented("DIGNOSTICS")
}

func (m Model) renderReporting() string {
	return m.renderNotImplemented("REPORTING VIEW")
}

func (m Model) renderSystem() string {
	doc := strings.Builder{}
	cpuInfo := fmt.Sprintf("CPU: %s %.1f%% | Load: %.2f, %.2f, %.2f", m.Monitor.CpuProgress.View(), m.Monitor.Cpu.Usage, m.Monitor.Cpu.LoadAvg1, m.Monitor.Cpu.LoadAvg5, m.Monitor.Cpu.LoadAvg15)
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("CPU"))
	doc.WriteString("\n")
	doc.WriteString(cpuInfo)
	doc.WriteString("\n\n")
	memInfo := fmt.Sprintf("RAM: %s %.1f%% | Total: %s | Used: %s | Free: %s", m.Monitor.MemProgress.View(), m.Monitor.Memory.Usage, utils.FormatBytes(m.Monitor.Memory.Total), utils.FormatBytes(m.Monitor.Memory.Used), utils.FormatBytes(m.Monitor.Memory.Free))
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Memory"))
	doc.WriteString("\n")
	doc.WriteString(memInfo)
	doc.WriteString("\n")
	swapInfo := fmt.Sprintf("SWP: %s %.1f%% | Total: %s | Used: %s | Free: %s", m.Monitor.SwapProgress.View(), m.Monitor.Memory.SwapUsage, utils.FormatBytes(m.Monitor.Memory.SwapTotal), utils.FormatBytes(m.Monitor.Memory.SwapUsed), utils.FormatBytes(m.Monitor.Memory.SwapFree))
	doc.WriteString(swapInfo)
	doc.WriteString("\n\n")
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Disks"))
	doc.WriteString("\n")
	for _, disk := range m.Monitor.Disks {
		if disk.Total > 0 {
			if p, ok := m.Monitor.DiskProgress[disk.Mountpoint]; ok {
				diskInfo := fmt.Sprintf("%-10s %s %.1f%% (%s/%s)", utils.Ellipsis(disk.Mountpoint, 10), p.View(), disk.Usage, utils.FormatBytes(disk.Used), utils.FormatBytes(disk.Total))
				doc.WriteString(diskInfo)
				doc.WriteString("\n")
			}
		}
	}
	return doc.String()
}

func (m Model) renderProcesses() string {
	p := "Search a process..."
	return m.renderTable(m.Monitor.ProcessTable, p)
}

func (m Model) interfaces() string {
	content := strings.Builder{}

	statusIcon := "🔴"
	statusText := "Disconnected"
	statusColor := errorColor

	if m.Network.NetworkResource.Connected {
		statusIcon = "🟢"
		statusText = "Connected"
		statusColor = successColor
	}

	statusLine := fmt.Sprintf("%s %s",
		statusIcon,
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText))
	content.WriteString(statusLine)

	if len(m.Network.NetworkResource.Interfaces) > 0 {
		tableContent := m.renderTable(m.Network.NetworkTable, "No network interfaces")
		content.WriteString(tableContent)
	}

	return cardNetworkStyle.
		Margin(0, 0, 0, 0).
		Padding(0, 1).
		Render(content.String())
}

func (m Model) renderNetwork() string {
	currentView := ""
	switch m.Network.SelectedItem {
	case model.NetworkTabInterface:
		currentView = m.interfaces()
	case model.NetworkTabConnectivity:
		currentView = m.renderNotImplemented("Connectivity Analysis")
	case model.NetworkTabConfiguration:
		currentView = m.renderNotImplemented("Network Configuration")
	case model.NetworkTabProtocol:
		currentView = m.renderNotImplemented("Protocol Analysis")
	}
	return currentView
}

func (m Model) renderNotImplemented(feature string) string {
	return cardStyle.Render(fmt.Sprintf("🚧 %s\n\nThis feature is not yet implemented.\n\nCheck back in future updates!", feature))
}

func (m Model) renderFooter() string {
	statusLine := ""
	if m.OperationInProgress {
		statusStyle := lipgloss.NewStyle().
			Foreground(clearWhite).
			Background(purpleCollor).
			Padding(0, 1).
			Bold(true)
		statusLine += statusStyle.Render("⏳ Operation in progress...") + "\n"
	} else if m.LastOperationMsg != "" {
		var statusStyle lipgloss.Style
		if strings.Contains(m.LastOperationMsg, "failed") || strings.Contains(m.LastOperationMsg, "Error") {
			statusStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Background(errorColor).
				Padding(0, 1).
				Bold(true)
		} else {
			statusStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Background(successColor).
				Padding(0, 1).
				Bold(true)
		}
		statusLine += statusStyle.Render(m.LastOperationMsg) + "\n"
	}

	var hints string
	switch m.Ui.State {
	case model.StateHome:
		hints = "[Enter] Select • [Tab/←→] Navigate • [1-4] Quick select • [q] Quit"
	case model.StateMonitor:
		hints = "[Enter] Select • [Tab/←→] Navigate • [1-3] Quick select • [b] Back • [q] Quit"
	case model.StateSystem:
		hints = "[↑↓] Scroll • [b] Back • [q] Quit"
	case model.StateProcess:
		hints = "[↑↓] Navigate • [/] Search • [k] Kill • [s/m] Sort • [b] Back • [q] Quit"
	case model.StateContainers:
		hints = "[↑↓] Navigate • [Enter] Menu • [/] Search • [b] Back • [q] Quit"
	case model.StateContainer:
		hints = "[Tab/←→] Switch tabs • [b] Back • [q] Quit"
	case model.StateContainerLogs:
		hints = "[↑↓] Scroll • [r] Refresh • [b] Back • [q] Quit"
	case model.StateNetwork:
		hints = "[Tab/←→] Switch tabs • [b] Back • [q] Quit"
	case model.StateDiagnostics, model.StateReporting:
		hints = "[b] Back • [q] Quit"
	}

	if m.ConfirmationVisible {
		hints = "[y] Confirm • [n/ESC] Cancel"
	} else if m.Monitor.ContainerMenuState == ContainerMenuVisible {
		hints = "[↑↓] Navigate • [Enter] Select • [ESC/b] Close • [o/l/...] Actions"
	}

	return statusLine + "\n" + hints
}

func (m Model) renderContainerMenu() string {
	if m.Monitor.ContainerMenuState != ContainerMenuVisible || m.Monitor.SelectedContainer == nil {
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

	return menuStyle.Render(doc.String())
}

func (m Model) renderContainerSingleView() string {
	if m.Monitor.SelectedContainer == nil {
		return cardStyle.Render("No container selected.")
	}
	style := lipgloss.NewStyle().Padding(0, 2).
		Foreground(whiteColor).
		Background(successColor).
		Bold(true)
	tabs := m.renderNav(m.Monitor.ContainerTabs, m.ContainerTab, style)

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
			m.getStatusWithIcon(m.Monitor.ContainerDetails.Status),
			m.Monitor.ContainerDetails.Project,
			m.Monitor.ContainerDetails.CreatedAt,
			m.Monitor.ContainerDetails.Uptime,
			m.getHealthWithIcon(m.Monitor.ContainerDetails.HealthCheck),
			m.Monitor.ContainerDetails.IPAddress,
			m.Monitor.ContainerDetails.Gateway,
		)
		doc.WriteString(metricLabelStyle.Render(info))

		if len(m.Monitor.ContainerDetails.Ports) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Ports:"))
			doc.WriteString("\n")
			for _, port := range m.Monitor.ContainerDetails.Ports {
				portInfo := fmt.Sprintf("  %s:%d → %d/%s", port.HostIP, port.PublicPort, port.PrivatePort, port.Type)
				doc.WriteString(metricValueStyle.Render(portInfo))
				doc.WriteString("\n")
			}
		}

		if m.Monitor.ContainerDetails.Command != "" {
			doc.WriteString("\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Command:"))
			doc.WriteString("\n")
			doc.WriteString(metricValueStyle.Render("  " + m.Monitor.ContainerDetails.Command))
		}

		if len(m.Monitor.ContainerDetails.Environment) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Environment:"))
			doc.WriteString(fmt.Sprintf("  %d variables", len(m.Monitor.ContainerDetails.Environment)))
		}
	} else {
		info := "Loading container details..."
		doc.WriteString(metricLabelStyle.Render(info))
	}

	return cardStyle.Render(doc.String())
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
		barColor := m.getCPUColor(cpuPercent)

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
		doc.WriteString(metricLabelStyle.Render("Loading CPU metrics..."))
	}

	return cardStyle.Render(doc.String())
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
		barColor := m.getMemoryColor(memPercent)

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
		doc.WriteString(metricLabelStyle.Render("Loading memory metrics..."))
	}

	return cardStyle.Render(doc.String())
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
			doc.WriteString(metricValueStyle.Render("  " + m.Monitor.ContainerDetails.IPAddress))
		}
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading network metrics..."))
	}

	return cardStyle.Render(doc.String())
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
		doc.WriteString(metricLabelStyle.Render("Loading disk metrics..."))
	}

	return cardStyle.Render(doc.String())
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
					metricLabelStyle.Render(key),
					metricValueStyle.Render(value)))
			}
		}
	} else if m.Monitor.ContainerDetails != nil {
		doc.WriteString(metricLabelStyle.Render("No environment variables found"))
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading environment variables..."))
	}

	return cardStyle.Render(doc.String())
}

func (m Model) getStatusWithIcon(status string) string {
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

func (m Model) getHealthWithIcon(health string) string {
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

func (m Model) getCPUColor(percent float64) lipgloss.Color {
	switch {
	case percent < 50:
		return lipgloss.Color("46") // Green
	case percent < 80:
		return lipgloss.Color("226") // Yellow
	default:
		return lipgloss.Color("196") // Red
	}
}

func (m Model) getMemoryColor(percent float64) lipgloss.Color {
	switch {
	case percent < 60:
		return lipgloss.Color("46") // Green
	case percent < 85:
		return lipgloss.Color("226") // Yellow
	default:
		return lipgloss.Color("196") // Red
	}
}

func (m Model) renderConfirmationDialog() string {
	if !m.ConfirmationVisible {
		return ""
	}

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render("⚠️  CONFIRMATION REQUIRED"))
	doc.WriteString("\n\n")

	doc.WriteString(metricLabelStyle.Render(m.ConfirmationMessage))
	doc.WriteString("\n\n")

	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Are you sure?"))
	doc.WriteString("\n")
	doc.WriteString("Press 'y' to confirm or 'n' to cancel")

	confirmationStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Padding(2).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255"))

	return confirmationStyle.Render(doc.String())
}

func (m Model) renderContainerLogs() string {
	if m.Monitor.SelectedContainer == nil {
		return cardStyle.Render("No container selected")
	}

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(fmt.Sprintf("Logs: %s", m.Monitor.SelectedContainer.Name)))
	doc.WriteString("\n\n")

	if m.Monitor.ContainerLogsLoading {
		doc.WriteString(metricLabelStyle.Render("Loading logs..."))
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
		doc.WriteString(metricLabelStyle.Render("No logs available or logs are empty"))
	}

	doc.WriteString("\n\n")
	doc.WriteString(metricLabelStyle.Render("Press 'r' to refresh logs, 'b' to go back"))

	return cardStyle.Render(doc.String())
}
