package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHome() string {
	var menu []string
	header := []string{}
	headerStyle := lipgloss.NewStyle().
		MarginLeft(5).
		Padding(0, 1).
		Foreground(lipgloss.Color("255")).
		Background(successColor).
		Bold(true)
	header = append(header, headerStyle.Render(asciiArt))

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

func (m Model) renderTabs() string {
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

func (m Model) renderMonitor() string {
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

func (m Model) renderApplications() string {
	return m.renderContainersTable()
}

// container
func (m Model) renderContainersTable() string {
	content := strings.Builder{}

	if m.searchMode {
		searchBar := searchBarStyle.
			Render(m.searchInput.View())
		content.WriteString(searchBar)
		content.WriteString("\n")
	}

	content.WriteString(m.container.View())

	return containerTableStyle.Render(content.String())
}

func (m Model) renderDignostics() string {
	return "DIGNOSTICS"
}

func (m Model) renderReporting() string {
	return "REPORTING VIEW"
}

func (m Model) renderSystem() string {
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

func (m Model) renderProcesses() string {
	return m.renderProcessTable()
}

func (m Model) renderProcessTable() string {
	content := strings.Builder{}

	if m.searchMode {
		searchBar := searchBarStyle.
			Render(m.searchInput.View())
		content.WriteString(searchBar)
		content.WriteString("\n")
	}

	content.WriteString(m.processTable.View())

	return cardStyle.Render(content.String())
}

func (m Model) renderNetwork() string {
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

func (m Model) renderFooter() string {
	footer := "\n"

	// Afficher le message de statut s'il y en a un
	if m.operationInProgress {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Padding(0, 1).
			Bold(true)
		footer += statusStyle.Render("⏳ Operation in progress...") + "\n"
	} else if m.lastOperationMsg != "" {
		var statusStyle lipgloss.Style
		if strings.Contains(m.lastOperationMsg, "failed") || strings.Contains(m.lastOperationMsg, "Error") {
			statusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(errorColor).
				Padding(0, 1).
				Bold(true)
		} else {
			statusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(successColor).
				Padding(0, 1).
				Bold(true)
		}
		footer += statusStyle.Render(m.lastOperationMsg) + "\n"
	}

	// Vue détaillée du conteneur
	if m.containerViewState == ContainerViewSingle {
		footer += "[b] Back to containers • [Tab/←→] Switch tabs • [1-6] Quick tab • [q] Quit"
		// Vue des logs du conteneur
	} else if m.containerViewState == ContainerViewLogs {
		footer += "[r] Refresh logs • [b] Back to containers • [q] Quit"
		// Boîte de confirmation
	} else if m.confirmationVisible {
		footer += "[y] Confirm • [n] Cancel • [ESC] Cancel"
		// Menu contextuel du conteneur
	} else if m.containerMenuState == ContainerMenuVisible {
		footer += "[↑↓] Navigate • [Enter] Select action • [ESC/b] Close menu • [o/l/r/d/x/s/p/e] Direct action"
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
func (m Model) renderContainerMenu() string {
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

	return menuStyle.Render(doc.String())
}

// Rendu de la vue détaillée du conteneur
func (m Model) renderContainerSingleView() string {
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
func (m Model) renderContainerGeneral() string {
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
		// Informations de base avec les détails complets (style ctop)
		info := fmt.Sprintf("ID: %s\nName: %s\nImage: %s\nStatus: %s\nProject: %s\nCreated: %s\nUptime: %s\nHealth: %s\nIP Address: %s\nGateway: %s",
			m.containerDetails.ID,
			m.containerDetails.Name,
			m.containerDetails.Image,
			m.getStatusWithIcon(m.containerDetails.Status),
			m.containerDetails.Project,
			m.containerDetails.CreatedAt,
			m.containerDetails.Uptime,
			m.getHealthWithIcon(m.containerDetails.HealthCheck),
			m.containerDetails.IPAddress,
			m.containerDetails.Gateway,
		)
		doc.WriteString(metricLabelStyle.Render(info))

		// Ports détaillés
		if len(m.containerDetails.Ports) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Ports:"))
			doc.WriteString("\n")
			for _, port := range m.containerDetails.Ports {
				portInfo := fmt.Sprintf("  %s:%d → %d/%s", port.HostIP, port.PublicPort, port.PrivatePort, port.Type)
				doc.WriteString(metricValueStyle.Render(portInfo))
				doc.WriteString("\n")
			}
		}

		// Command and entrypoint (like ctop)
		if m.containerDetails.Command != "" {
			doc.WriteString("\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Command:"))
			doc.WriteString("\n")
			doc.WriteString(metricValueStyle.Render("  " + m.containerDetails.Command))
		}

		// Environment variables count
		if len(m.containerDetails.Environment) > 0 {
			doc.WriteString("\n\n")
			doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Environment:"))
			doc.WriteString(fmt.Sprintf("  %d variables", len(m.containerDetails.Environment)))
		}
	} else {
		// Informations de base minimales
		info := "Loading container details..."
		doc.WriteString(metricLabelStyle.Render(info))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation CPU du conteneur
func (m Model) renderContainerCPU() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("CPU Usage"))
	doc.WriteString("\n\n")

	if m.containerDetails != nil {
		// Informations CPU détaillées (style ctop)
		cpuPercent := m.containerDetails.Stats.CPUPercent
		doc.WriteString(fmt.Sprintf("Usage: %.1f%%\n", cpuPercent))

		// Barre de progression colorée
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

		// Graphique en temps réel
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Usage History:"))
		doc.WriteString("\n")
		chart := m.renderCPUChart(50, 10)
		doc.WriteString(chart)
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading CPU metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation mémoire du conteneur
func (m Model) renderContainerMemory() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Memory Usage"))
	doc.WriteString("\n\n")

	if m.containerDetails != nil {
		// Informations mémoire détaillées (style ctop)
		memPercent := m.containerDetails.Stats.MemoryPercent
		memUsage := m.containerDetails.Stats.MemoryUsage
		memLimit := m.containerDetails.Stats.MemoryLimit

		doc.WriteString(fmt.Sprintf("Usage: %.1f%%\n", memPercent))
		doc.WriteString(fmt.Sprintf("Used: %s\n", utils.FormatBytes(memUsage)))
		doc.WriteString(fmt.Sprintf("Limit: %s\n", utils.FormatBytes(memLimit)))
		doc.WriteString(fmt.Sprintf("Available: %s\n\n", utils.FormatBytes(memLimit-memUsage)))

		// Barre de progression colorée
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

		// Graphique en temps réel
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Usage History:"))
		doc.WriteString("\n")
		chart := m.renderMemoryChart(50, 10)
		doc.WriteString(chart)
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading memory metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation réseau du conteneur
func (m Model) renderContainerNetwork() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Network Usage"))
	doc.WriteString("\n\n")

	if m.containerDetails != nil {
		// Statistiques réseau détaillées (style ctop)
		rxBytes := m.containerDetails.Stats.NetworkRx
		txBytes := m.containerDetails.Stats.NetworkTx

		doc.WriteString(fmt.Sprintf("RX: %s/s\n", utils.FormatBytes(rxBytes)))
		doc.WriteString(fmt.Sprintf("TX: %s/s\n\n", utils.FormatBytes(txBytes)))

		// Graphiques réseau avec échelle en MB/s
		doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Receive Traffic (MB/s):"))
		doc.WriteString("\n")
		rxChart := m.renderNetworkRXChart(50, 6)
		doc.WriteString(rxChart)

		doc.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("Transmit Traffic (MB/s):"))
		doc.WriteString("\n")
		txChart := m.renderNetworkTXChart(50, 6)
		doc.WriteString(txChart)

		// Informations réseau supplémentaires si disponibles
		if m.containerDetails.IPAddress != "" {
			doc.WriteString("\n\n" + lipgloss.NewStyle().Bold(true).Render("Network Interfaces:"))
			doc.WriteString("\n")
			doc.WriteString(metricValueStyle.Render("  " + m.containerDetails.IPAddress))
		}
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading network metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu de l'utilisation disque du conteneur
func (m Model) renderContainerDisk() string {
	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render("Disk I/O"))
	doc.WriteString("\n\n")

	if m.containerDetails != nil {
		// Statistiques disque détaillées (style ctop)
		readBytes := m.containerDetails.Stats.BlockRead
		writeBytes := m.containerDetails.Stats.BlockWrite

		doc.WriteString(fmt.Sprintf("Read: %s\n", utils.FormatBytes(readBytes)))
		doc.WriteString(fmt.Sprintf("Write: %s\n\n", utils.FormatBytes(writeBytes)))

		// Graphiques en barres avec échelle relative
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

		// Graphiques d'historique
		doc.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("I/O History:"))
		doc.WriteString("\n")
		// Note: Les graphiques d'historique disque seront ajoutés ultérieurement
		doc.WriteString("Disk I/O history charts coming soon...")
	} else {
		doc.WriteString(metricLabelStyle.Render("Loading disk metrics..."))
	}

	return cardStyle.Render(doc.String())
}

// Rendu des variables d'environnement du conteneur
func (m Model) renderContainerEnv() string {
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

// Helper methods for ctop-like functionality

// getStatusWithIcon returns status with appropriate icon like ctop
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

// getHealthWithIcon returns health status with appropriate icon like ctop
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

// getCPUColor returns color based on CPU usage percentage (like ctop)
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

// getMemoryColor returns color based on memory usage percentage (like ctop)
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

// Rendu de la boîte de confirmation
func (m Model) renderConfirmationDialog() string {
	if !m.confirmationVisible {
		return ""
	}

	doc := strings.Builder{}

	// Titre de confirmation
	doc.WriteString(lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render("⚠️  CONFIRMATION REQUIRED"))
	doc.WriteString("\n\n")

	// Message de confirmation
	doc.WriteString(metricLabelStyle.Render(m.confirmationMessage))
	doc.WriteString("\n\n")

	// Options
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Are you sure?"))
	doc.WriteString("\n")
	doc.WriteString("Press 'y' to confirm or 'n' to cancel")

	// Style de la boîte de dialogue
	confirmationStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Padding(2).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255"))

	return confirmationStyle.Render(doc.String())
}

// Rendu de la vue des logs du conteneur
func (m Model) renderContainerLogs() string {
	if m.selectedContainer == nil {
		return cardStyle.Render("No container selected")
	}

	doc := strings.Builder{}

	doc.WriteString(lipgloss.NewStyle().Bold(true).Underline(true).MarginBottom(1).Render(fmt.Sprintf("Logs: %s", m.selectedContainer.Name)))
	doc.WriteString("\n\n")

	if m.containerLogsLoading {
		doc.WriteString(metricLabelStyle.Render("Loading logs..."))
	} else if m.containerLogs != "" {
		// Afficher les logs dans une zone scrollable
		logStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("235")).
			Padding(1).
			Height(20).
			Width(80)

		doc.WriteString(logStyle.Render(m.containerLogs))
	} else {
		doc.WriteString(metricLabelStyle.Render("No logs available or logs are empty"))
	}

	doc.WriteString("\n\n")
	doc.WriteString(metricLabelStyle.Render("Press 'r' to refresh logs, 'b' to go back"))

	return cardStyle.Render(doc.String())
}
