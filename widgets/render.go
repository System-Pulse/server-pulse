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

	// Vue détaillée du conteneur
	if m.containerViewState == ContainerViewSingle {
		footer += "[b] Back to containers • [Tab/←→] Switch tabs • [1-6] Quick tab • [q] Quit"
		// Menu contextuel du conteneur
	} else if m.containerMenuState == ContainerMenuVisible {
		footer += "[↑↓] Navigate • [Enter] Select action • [ESC/b] Close menu • [o/l/r/d/s/p/e] Direct action"
	} else if m.activeView == -1 {
		footer += "[Enter] Select view • [q] Quit • [Tab/←→] Navigate • [1-4] Quick select"
	} else if m.isMonitorActive {
		footer += "[b] Back • [Tab/←→] Switch • [/] Search • [q] Quit"
		switch m.selectedMonitor {
		case 1:
			footer += " • [↑↓] Navigate • [k] Kill • [s] Sort CPU • [m] Sort Mem"
		case 2:
			footer += " • [↑↓] Navigate • [Enter] Container menu"
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

// Rendu du menu contextuel des conteneurs
func (m model) renderContainerMenu() string {
	if m.containerMenuState != ContainerMenuVisible || m.selectedContainer == nil {
		return ""
	}

	doc := strings.Builder{}

	// Titre simple
	doc.WriteString("CONTAINER MENU\n")
	doc.WriteString(fmt.Sprintf("Container: %s\n", m.selectedContainer.Name))
	doc.WriteString(fmt.Sprintf("Status: %s\n", m.selectedContainer.Status))
	doc.WriteString("\n")

	// Options du menu
	for i, item := range m.containerMenuItems {
		prefix := "  "
		if i == m.selectedMenuItem {
			prefix = "> "
		}
		doc.WriteString(fmt.Sprintf("%s[%s] %s\n", prefix, item.Key, item.Label))
	}

	doc.WriteString("\n")
	doc.WriteString("Navigation: ↑↓ Navigate • Enter Select • ESC Close\n")

	// Style simple avec bordure
	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("57")).
		Padding(1).
		Background(lipgloss.Color("235"))

	return menuStyle.Render(doc.String())
}

// Rendu de la vue détaillée du conteneur
func (m model) renderContainerSingleView() string {
	if m.containerViewState != ContainerViewSingle || m.selectedContainer == nil {
		return ""
	}

	// En-tête avec les onglets
	var tabs []string
	for i, tab := range m.containerTabs {
		style := lipgloss.NewStyle().Padding(0, 2)
		if ContainerTab(i) == m.containerTab {
			style = style.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(true)
		} else {
			style = style.
				Foreground(lipgloss.Color("240")).
				Background(lipgloss.Color("236"))
		}
		tabs = append(tabs, style.Render(tab))
	}

	tabsHeader := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Contenu selon l'onglet sélectionné
	var content string
	switch m.containerTab {
	case ContainerTabGeneral:
		content = m.renderContainerGeneral()
	case ContainerTabCPU:
		content = m.renderContainerCPU()
	case ContainerTabMemory:
		content = m.renderContainerMemory()
	case ContainerTabNetwork:
		content = m.renderContainerNetwork()
	case ContainerTabDisk:
		content = m.renderContainerDisk()
	case ContainerTabEnv:
		content = m.renderContainerEnv()
	default:
		content = m.renderContainerGeneral()
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabsHeader, content)
}

// Rendu des informations générales du conteneur
func (m model) renderContainerGeneral() string {
	doc := strings.Builder{}

	containerName := "N/A"
	if m.containerDetails != nil {
		containerName = m.containerDetails.Name
	} else if m.selectedContainer != nil {
		containerName = m.selectedContainer.Name
	}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(fmt.Sprintf("Container: %s", containerName)))
	doc.WriteString("\n\n")

	if m.containerDetails != nil {
		// Informations de base avec les détails complets
		info := fmt.Sprintf("ID: %s\nName: %s\nImage: %s\nStatus: %s\nProject: %s\nCreated: %s\nUptime: %s\nHealth: %s\nIP Address: %s\nGateway: %s",
			m.containerDetails.ID,
			m.containerDetails.Name,
			m.containerDetails.Image,
			m.containerDetails.Status,
			m.containerDetails.Project,
			m.containerDetails.CreatedAt,
			m.containerDetails.Uptime,
			m.containerDetails.HealthCheck,
			m.containerDetails.IPAddress,
			m.containerDetails.Gateway,
		)
		doc.WriteString(metricLabelStyle.Render(info))

		// Ports
		if len(m.containerDetails.Ports) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Ports:"))
			doc.WriteString("\n")
			for _, port := range m.containerDetails.Ports {
				portInfo := fmt.Sprintf("  %d:%d/%s", port.PublicPort, port.PrivatePort, port.Type)
				doc.WriteString(metricValueStyle.Render(portInfo))
				doc.WriteString("\n")
			}
		}
	} else {
		// Informations de base minimales
		info := "Loading container details..."
		doc.WriteString(metricLabelStyle.Render(info))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation CPU du conteneur
func (m model) renderContainerCPU() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("CPU Usage"))
	doc.WriteString("\n\n")
	/*
		if m.containerDetails != nil {
			doc.WriteString(metricLabelStyle.Render("CPU Usage: "))
			doc.WriteString(metricValueStyle.Render(fmt.Sprintf("%.2f%%", m.containerDetails.Stats.CPUPercent)))
			doc.WriteString("\n\n")

			// Graphique simple de l'utilisation CPU
			cpuBar := ""
			cpuBlocks := int(m.containerDetails.Stats.CPUPercent / 5) // 20 blocs max (100/5)
			for i := range 20 {
				if i < cpuBlocks {
					cpuBar += "█"
				} else {
					cpuBar += "░"
				}
			}
			doc.WriteString(fmt.Sprintf("CPU [%s] %.1f%%\n", cpuBar, m.containerDetails.Stats.CPUPercent))
		} else {
			doc.WriteString(metricLabelStyle.Render("Loading CPU metrics..."))
		}*/

	if m.containerDetails != nil {
		// Barre de progression
		cpuBar := ""
		cpuBlocks := int(m.containerDetails.Stats.CPUPercent / 5)
		for i := 0; i < 20; i++ {
			if i < cpuBlocks {
				cpuBar += "█"
			} else {
				cpuBar += "░"
			}
		}
		doc.WriteString(fmt.Sprintf("CPU Usage: [%s] %.1f%%\n\n", cpuBar, m.containerDetails.Stats.CPUPercent))

		// Graphique en temps réel
		chart := m.renderCPUChart(50, 12)
		doc.WriteString(chart)
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading CPU metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation mémoire du conteneur
func (m model) renderContainerMemory() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Memory Usage"))
	doc.WriteString("\n\n")

	/*if m.containerDetails != nil {
		doc.WriteString(metricLabelStyle.Render("Memory Usage: "))
		doc.WriteString(metricValueStyle.Render(fmt.Sprintf("%s / %s (%.1f%%)",
			utils.FormatBytes(m.containerDetails.Stats.MemoryUsage),
			utils.FormatBytes(m.containerDetails.Stats.MemoryLimit),
			m.containerDetails.Stats.MemoryPercent)))
		doc.WriteString("\n\n")

		// Graphique simple de l'utilisation mémoire
		memBar := ""
		memBlocks := int(m.containerDetails.Stats.MemoryPercent / 5) // 20 blocs max
		for i := range 20 {
			if i < memBlocks {
				memBar += "█"
			} else {
				memBar += "░"
			}
		}
		doc.WriteString(fmt.Sprintf("MEM [%s] %.1f%%\n", memBar, m.containerDetails.Stats.MemoryPercent))
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading memory metrics..."))
	}*/

	if m.containerDetails != nil {
		// Barre de progression
		memBar := ""
		memBlocks := int(m.containerDetails.Stats.MemoryPercent / 5)
		for i := 0; i < 20; i++ {
			if i < memBlocks {
				memBar += "█"
			} else {
				memBar += "░"
			}
		}
		doc.WriteString(fmt.Sprintf("Memory: [%s] %.1f%%\n", memBar, m.containerDetails.Stats.MemoryPercent))
		doc.WriteString(fmt.Sprintf("Usage: %s / %s\n\n",
			utils.FormatBytes(m.containerDetails.Stats.MemoryUsage),
			utils.FormatBytes(m.containerDetails.Stats.MemoryLimit)))

		// Graphique en temps réel
		chart := m.renderMemoryChart(50, 12)
		doc.WriteString(chart)
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading memory metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation réseau du conteneur
func (m model) renderContainerNetwork() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Network Usage"))
	doc.WriteString("\n\n")

	/*if m.containerDetails != nil {
		doc.WriteString(metricLabelStyle.Render("RX Bytes: "))
		doc.WriteString(metricValueStyle.Render(utils.FormatBytes(m.containerDetails.Stats.NetworkRx)))
		doc.WriteString("\n")
		doc.WriteString(metricLabelStyle.Render("TX Bytes: "))
		doc.WriteString(metricValueStyle.Render(utils.FormatBytes(m.containerDetails.Stats.NetworkTx)))
		doc.WriteString("\n\n")

		// Graphiques en barres simples pour RX/TX
		maxBytes := max(m.containerDetails.Stats.NetworkTx, m.containerDetails.Stats.NetworkRx)

		if maxBytes > 0 {
			rxBlocks := int((float64(m.containerDetails.Stats.NetworkRx) / float64(maxBytes)) * 20)
			txBlocks := int((float64(m.containerDetails.Stats.NetworkTx) / float64(maxBytes)) * 20)

			rxBar := strings.Repeat("█", rxBlocks) + strings.Repeat("░", 20-rxBlocks)
			txBar := strings.Repeat("█", txBlocks) + strings.Repeat("░", 20-txBlocks)

			doc.WriteString(fmt.Sprintf("RX [%s]\n", rxBar))
			doc.WriteString(fmt.Sprintf("TX [%s]\n", txBar))
		}
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading network metrics..."))
	}*/
	if m.containerDetails != nil {
		doc.WriteString(fmt.Sprintf("RX: %s/s | TX: %s/s\n\n",
			utils.FormatBytes(m.containerDetails.Stats.NetworkRx),
			utils.FormatBytes(m.containerDetails.Stats.NetworkTx)))

		// Graphiques réseau
		doc.WriteString("RX Traffic:\n")
		rxChart := m.renderNetworkRXChart(50, 8)
		doc.WriteString(rxChart)
		doc.WriteString("\n\nTX Traffic:\n")
		txChart := m.renderNetworkTXChart(50, 8)
		doc.WriteString(txChart)
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading network metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation disque du conteneur
func (m model) renderContainerDisk() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Disk Usage"))
	doc.WriteString("\n\n")

	if m.containerDetails != nil {
		doc.WriteString(metricLabelStyle.Render("Disk Read: "))
		doc.WriteString(metricValueStyle.Render(utils.FormatBytes(m.containerDetails.Stats.BlockRead)))
		doc.WriteString("\n")
		doc.WriteString(metricLabelStyle.Render("Disk Write: "))
		doc.WriteString(metricValueStyle.Render(utils.FormatBytes(m.containerDetails.Stats.BlockWrite)))
		doc.WriteString("\n\n")

		// Graphiques en barres simples pour Read/Write
		maxBytes := max(m.containerDetails.Stats.BlockWrite, m.containerDetails.Stats.BlockRead)

		if maxBytes > 0 {
			readBlocks := int((float64(m.containerDetails.Stats.BlockRead) / float64(maxBytes)) * 20)
			writeBlocks := int((float64(m.containerDetails.Stats.BlockWrite) / float64(maxBytes)) * 20)

			readBar := strings.Repeat("█", readBlocks) + strings.Repeat("░", 20-readBlocks)
			writeBar := strings.Repeat("█", writeBlocks) + strings.Repeat("░", 20-writeBlocks)

			doc.WriteString(fmt.Sprintf("READ  [%s]\n", readBar))
			doc.WriteString(fmt.Sprintf("WRITE [%s]\n", writeBar))
		}
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading disk metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu des variables d'environnement du conteneur
func (m model) renderContainerEnv() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Environment Variables"))
	doc.WriteString("\n\n")

	if m.containerDetails != nil && len(m.containerDetails.Environment) > 0 {
		for _, env := range m.containerDetails.Environment {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				// Limiter la longueur de la valeur pour l'affichage
				if len(value) > 50 {
					value = value[:47] + "..."
				}
				doc.WriteString(fmt.Sprintf("%s: %s\n",
					metricLabelStyle.Render(key),
					metricValueStyle.Render(value)))
			}
		}
	} else if m.containerDetails != nil {
		doc.WriteString(metricLabelStyle.Render("No environment variables found"))
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading environment variables..."))
	}

	return cardStyle.Render(doc.String())
}
