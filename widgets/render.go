package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/lipgloss"
)

func (m model) renderHome() string {
	var menu []string
	header := []string{}
	headerStyle := lipgloss.NewStyle().
		MarginLeft(5).
		Padding(0, 1).
		Foreground(lipgloss.Color("255")).
		Background(successColor).
		Bold(true)
	header = append(header, headerStyle.Render("Server-Pulse"))

	for i, t := range m.tabs.DashBoard {
		style := lipgloss.NewStyle()
		if i == m.selectedTab {
			if m.activeView == i {
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
	systemInfo := fmt.Sprintf("Host: %s\nOS: %s\nKernel: %s\nUptime: %s", m.system.Hostname, m.system.OS, m.system.Kernel, utils.FormatUptime(m.system.Uptime))
	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("System Info:"))
	doc.WriteString("\n")
	doc.WriteString(metricLabelStyle.Render(systemInfo))
	doc.WriteString("\n")
	header = append(header, cardStyle.MarginBottom(0).Render(doc.String()))

	header = append(header, lipgloss.JoinHorizontal(lipgloss.Top, menu...))
	return lipgloss.JoinVertical(lipgloss.Top, header...)
}

func (m model) renderTabs() string {
	if m.isMonitorActive {
		var tabs []string
		for i, t := range m.tabs.Monitor {
			style := lipgloss.NewStyle().Padding(0, 1)
			if i == m.selectedMonitor {
				style = style.
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57")).
					Bold(true)
			} else {
				style = style.
					Foreground(lipgloss.Color("240")).
					Background(lipgloss.Color("236"))
			}
			tabs = append(tabs, style.Render(t))
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	}
	return ""
}

func (m model) renderMonitor() string {
	var currentView string
	switch m.selectedMonitor {
	case 0:
		currentView = m.renderSystem()
	case 1:
		currentView = m.renderProcesses()
	case 2:
		currentView = m.renderApplications()
	}
	return currentView
}

func (m model) renderApplications() string {
	return m.renderContainersTable()
}

// container
func (m model) renderContainersTable() string {
	content := strings.Builder{}

	if m.searchMode {
		searchBar := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(0, 1).
			MarginBottom(1).
			Render(m.searchInput.View())
		content.WriteString(searchBar)
		content.WriteString("\n")
	}

	content.WriteString(m.container.View())

	return cardStyle.Render(content.String())
}

func (m model) renderDignostics() string {
	return "DIGNOSTICS"
}

func (m model) renderReporting() string {
	return "REPORTING VIEW"
}

func (m model) renderSystem() string {
	doc := strings.Builder{}
	cpuInfo := fmt.Sprintf("CPU: %s %.1f%% | Load: %.2f, %.2f, %.2f", m.cpuProgress.View(), m.cpu.Usage, m.cpu.LoadAvg1, m.cpu.LoadAvg5, m.cpu.LoadAvg15)
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("CPU"))
	doc.WriteString("\n")
	doc.WriteString(cpuInfo)
	doc.WriteString("\n\n")
	memInfo := fmt.Sprintf("RAM: %s %.1f%% | Total: %s | Used: %s | Free: %s", m.memProgress.View(), m.memory.Usage, utils.FormatBytes(m.memory.Total), utils.FormatBytes(m.memory.Used), utils.FormatBytes(m.memory.Free))
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Memory"))
	doc.WriteString("\n")
	doc.WriteString(memInfo)
	doc.WriteString("\n")
	swapInfo := fmt.Sprintf("SWP: %s %.1f%% | Total: %s | Used: %s | Free: %s", m.swapProgress.View(), m.memory.SwapUsage, utils.FormatBytes(m.memory.SwapTotal), utils.FormatBytes(m.memory.SwapUsed), utils.FormatBytes(m.memory.SwapFree))
	doc.WriteString(swapInfo)
	doc.WriteString("\n\n")
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Disks"))
	doc.WriteString("\n")
	for _, disk := range m.disks {
		if disk.Total > 0 {
			if p, ok := m.diskProgress[disk.Mountpoint]; ok {
				diskInfo := fmt.Sprintf("%-10s %s %.1f%% (%s/%s)", utils.Ellipsis(disk.Mountpoint, 10), p.View(), disk.Usage, utils.FormatBytes(disk.Used), utils.FormatBytes(disk.Total))
				doc.WriteString(diskInfo)
				doc.WriteString("\n")
			}
		}
	}
	return doc.String()
}

func (m model) renderProcesses() string {
	return m.renderProcessTable()
}

func (m model) renderProcessTable() string {
	content := strings.Builder{}

	if m.searchMode {
		searchBar := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(0, 1).
			MarginBottom(1).
			Render(m.searchInput.View())
		content.WriteString(searchBar)
		content.WriteString("\n")
	}

	content.WriteString(m.processTable.View())

	return cardStyle.Render(content.String())
}

func (m model) renderNetwork() string {
	content := strings.Builder{}

	statusIcon := "🔴"
	statusText := "Disconnected"
	statusColor := errorColor

	if m.network.Connected {
		statusIcon = "🟢"
		statusText = "Connected"
		statusColor = successColor
	}

	content.WriteString(metricLabelStyle.Render("🌐 Network Status"))
	content.WriteString("\n\n")

	statusLine := fmt.Sprintf("%s %s",
		statusIcon,
		lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText))
	content.WriteString(statusLine)
	content.WriteString("\n\n")

	if len(m.network.PrivateIPs) > 0 {
		content.WriteString(metricLabelStyle.Render("Private IPs:"))
		content.WriteString("\n")
		for _, ip := range m.network.PrivateIPs {
			content.WriteString("  • " + metricValueStyle.Render(ip))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%-12s %s",
		metricLabelStyle.Render("Public IPv4:"),
		metricValueStyle.Render(m.network.PublicIPv4)))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%-12s %s",
		metricLabelStyle.Render("Public IPv6:"),
		metricValueStyle.Render(m.network.PublicIPv6)))

	return cardStyle.Render(content.String())
}

func (m model) renderFooter() string {
	footer := "\n"

	if m.activeView == -1 {
		footer += "[Enter] Select view • [q] Quit • [Tab/←→] Navigate • [1-4] Quick select"
	} else if m.isMonitorActive {
		footer += "[b] Back • [Tab/←→] Switch • / Search • [q] Quit"
		if m.selectedMonitor == 1 {
			footer += " • [↑↓] Navigate • [k] Kill • [s] Sort CPU • [m] Sort Mem"
		}
	} else {
		switch m.activeView {
		case 0: // Monitor
			footer += "[b] Back • [Enter] Select sub-menu • [q] Quit"
		case 1: // Diagnostic
			footer += "[b] Back • [q] Quit"
		case 2: // Network
			footer += "[b] Back • [q] Quit"
		case 3: // Reporting
			footer += "[b] Back • [q] Quit"
		}
	}
	return footer
}
