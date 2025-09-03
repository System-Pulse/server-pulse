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

func (m model) renderContainerMenu() string {
	if !m.containerMenu.Visible {
		return ""
	}

	menuItems := []string{
		"[o] Open single view",
		"[l] View container logs",
		"[r] Restart container",
		"[d] Delete container",
	}

	// Ajouter les options conditionnelles selon l'état
	if strings.Contains(strings.ToLower(m.containerMenu.Container.Status), "running") {
		menuItems = append(menuItems, "[s] Stop container")
		menuItems = append(menuItems, "[p] Pause container")
	} else if strings.Contains(strings.ToLower(m.containerMenu.Container.Status), "stopped") {
		menuItems = append(menuItems, "[s] Start container")
	} else if strings.Contains(strings.ToLower(m.containerMenu.Container.Status), "paused") {
		menuItems = append(menuItems, "[u] Unpause container")
	}

	menuItems = append(menuItems, "[e] Exec shell")
	menuItems = append(menuItems, "[esc] Close menu")

	var renderedItems []string
	for i, item := range menuItems {
		if i == m.containerMenu.Selected {
			renderedItems = append(renderedItems, selectedMenuItemStyle.Render("➤ "+item))
		} else {
			renderedItems = append(renderedItems, menuItemStyle.Render("  "+item))
		}
	}

	// Ajouter un titre au menu
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229")).
		// Background(accentColor).
		Padding(0, 2).
		Render(fmt.Sprintf(" Container: %s ", m.containerMenu.Container.Name))

	menuContent := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		strings.Join(renderedItems, "\n"),
	)

	return menuStyle.
		Width(30). // Largeur fixe pour le menu
		Render(menuContent)
}

func (m model) renderContainerSingleView() string {
	if !m.containerSingleView.Visible {
		return ""
	}

	doc := strings.Builder{}

	// Header avec le nom du conteneur
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("229")).
		Background(accentColor).
		Padding(1, 2).
		Width(m.width - 4).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("Container: %s", m.containerSingleView.Container.Name))

	doc.WriteString(header)
	doc.WriteString("\n\n")

	// Onglets
	var tabs []string
	for i, tab := range m.containerSingleView.Tabs {
		style := lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1)

		if i == m.containerSingleView.ActiveTab {
			style = style.
				Foreground(lipgloss.Color("229")).
				Background(accentColor).
				Bold(true)
		} else {
			style = style.
				Foreground(lipgloss.Color("240")).
				Background(lipgloss.Color("236"))
		}
		tabs = append(tabs, style.Render(tab))
	}

	tabsLine := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	doc.WriteString(tabsLine)
	doc.WriteString("\n\n")

	// Contenu selon l'onglet actif
	var content string
	switch m.containerSingleView.ActiveTab {
	case 0: // INFO
		content = m.renderContainerInfo()
	case 1: // CPU
		content = m.renderContainerCPU()
	case 2: // MEM
		content = m.renderContainerMemory()
	case 3: // NET
		content = m.renderContainerNetwork()
	case 4: // DISK
		content = m.renderContainerDisk()
	case 5: // ENV
		content = m.renderContainerEnv()
	}

	contentBox := cardStyle.
		Width(m.width - 6).
		Height(m.height - 12).
		Render(content)

	doc.WriteString(contentBox)

	// Footer avec informations contextuelles
	currentTab := m.containerSingleView.Tabs[m.containerSingleView.ActiveTab]
	containerState := m.containerSingleView.Stats.State

	footerText := fmt.Sprintf("[Tab/←→] Switch tab • [1-6] Direct access • [Esc/b] Back • [q] Quit | Tab: %s | State: %s",
		currentTab, containerState)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(m.width - 4).
		MarginTop(1).
		Render(footerText)

	doc.WriteString("\n")
	doc.WriteString(footer)

	return doc.String()
}

func (m model) renderContainerInfo() string {
	doc := strings.Builder{}

	stats := m.containerSingleView.Stats

	// Informations générales
	doc.WriteString(metricLabelStyle.Render("📋 General Information"))
	doc.WriteString("\n\n")

	info := [][]string{
		{"ID:", stats.ID},
		{"Name:", stats.Name},
		{"Image:", stats.Image},
		{"Ports:", stats.Ports},
		{"IPs:", strings.Join(stats.IPs, ", ")},
		{"State:", stats.State},
		{"Created:", stats.Created},
		{"Uptime:", stats.Uptime},
		{"Health:", stats.Health},
	}

	for _, row := range info {
		if row[1] != "" {
			line := fmt.Sprintf("%-12s %s",
				metricLabelStyle.Render(row[0]),
				metricValueStyle.Render(row[1]))
			doc.WriteString(line)
			doc.WriteString("\n")
		}
	}

	return doc.String()
}

func (m model) renderContainerCPU() string {
	doc := strings.Builder{}

	doc.WriteString(metricLabelStyle.Render("📊 CPU Usage"))
	doc.WriteString("\n\n")

	// Graphique ASCII simple pour l'utilisation CPU
	cpuPercent := m.containerSingleView.Stats.CPUUsage
	doc.WriteString(fmt.Sprintf("Current: %.2f%%\n\n", cpuPercent))

	// Graphique en barres horizontal avec couleurs
	barWidth := 50
	filled := int(cpuPercent / 100 * float64(barWidth))

	// Choisir la couleur selon le niveau d'utilisation
	var barColor lipgloss.Color
	switch {
	case cpuPercent < 25:
		barColor = successColor
	case cpuPercent < 50:
		barColor = lipgloss.Color("#fbbf24") // jaune
	case cpuPercent < 75:
		barColor = lipgloss.Color("#f97316") // orange
	default:
		barColor = errorColor
	}

	filledBar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled))
	emptyBar := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", barWidth-filled))

	doc.WriteString(fmt.Sprintf("[%s%s] %.1f%%\n\n", filledBar, emptyBar, cpuPercent))

	// Historique simple (simulation avec les données stockées)
	doc.WriteString("Recent History:\n")
	for i, val := range m.containerSingleView.CPUData[len(m.containerSingleView.CPUData)-10:] {
		doc.WriteString(fmt.Sprintf("-%ds: %.1f%%\n", (10-i)*2, val))
	}

	return doc.String()
}

func (m model) renderContainerMemory() string {
	doc := strings.Builder{}

	doc.WriteString(metricLabelStyle.Render("🧠 Memory Usage"))
	doc.WriteString("\n\n")

	memPercent := m.containerSingleView.Stats.MemUsage
	memLimit := m.containerSingleView.Stats.MemLimit
	memUsed := uint64(float64(memLimit) * memPercent / 100)

	doc.WriteString(fmt.Sprintf("Used: %s / %s (%.1f%%)\n\n",
		utils.FormatBytes(memUsed),
		utils.FormatBytes(memLimit),
		memPercent))

	// Graphique en barres avec couleurs
	barWidth := 50
	filled := int(memPercent / 100 * float64(barWidth))

	// Choisir la couleur selon le niveau d'utilisation mémoire
	var barColor lipgloss.Color
	switch {
	case memPercent < 30:
		barColor = successColor
	case memPercent < 60:
		barColor = lipgloss.Color("#fbbf24") // jaune
	case memPercent < 80:
		barColor = lipgloss.Color("#f97316") // orange
	default:
		barColor = errorColor
	}

	filledBar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled))
	emptyBar := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", barWidth-filled))

	doc.WriteString(fmt.Sprintf("[%s%s] %.1f%%\n\n", filledBar, emptyBar, memPercent))

	// Historique
	doc.WriteString("Recent History:\n")
	for i, val := range m.containerSingleView.MemoryData[len(m.containerSingleView.MemoryData)-10:] {
		doc.WriteString(fmt.Sprintf("-%ds: %.1f%%\n", (10-i)*2, val))
	}

	return doc.String()
}

func (m model) renderContainerNetwork() string {
	doc := strings.Builder{}

	doc.WriteString(metricLabelStyle.Render("🌐 Network Usage"))
	doc.WriteString("\n\n")

	netRX := m.containerSingleView.Stats.NetRX
	netTX := m.containerSingleView.Stats.NetTX

	doc.WriteString(fmt.Sprintf("RX: %s\n", utils.FormatBytes(netRX)))
	doc.WriteString(fmt.Sprintf("TX: %s\n\n", utils.FormatBytes(netTX)))

	// Graphique RX
	doc.WriteString("RX Traffic:\n")
	maxRX := float64(1024 * 1024) // 1MB max pour l'échelle
	if float64(netRX) > maxRX {
		maxRX = float64(netRX)
	}
	rxPercent := float64(netRX) / maxRX * 100

	barWidth := 30
	filled := int(rxPercent / 100 * float64(barWidth))
	filledBar := lipgloss.NewStyle().Foreground(accentColor).Render(strings.Repeat("█", filled))
	emptyBar := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", barWidth-filled))
	doc.WriteString(fmt.Sprintf("[%s%s] %s\n\n", filledBar, emptyBar, utils.FormatBytes(netRX)))

	// Graphique TX
	doc.WriteString("TX Traffic:\n")
	maxTX := float64(1024 * 1024) // 1MB max pour l'échelle
	if float64(netTX) > maxTX {
		maxTX = float64(netTX)
	}
	txPercent := float64(netTX) / maxTX * 100

	filled = int(txPercent / 100 * float64(barWidth))
	filledBar = lipgloss.NewStyle().Foreground(lipgloss.Color("#8b5cf6")).Render(strings.Repeat("█", filled)) // violet pour TX
	emptyBar = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", barWidth-filled))
	doc.WriteString(fmt.Sprintf("[%s%s] %s\n", filledBar, emptyBar, utils.FormatBytes(netTX)))

	return doc.String()
}

func (m model) renderContainerDisk() string {
	doc := strings.Builder{}

	doc.WriteString(metricLabelStyle.Render("💽 Disk Usage"))
	doc.WriteString("\n\n")

	diskUsage := m.containerSingleView.Stats.DiskUsage
	doc.WriteString(fmt.Sprintf("Disk Usage: %s\n\n", utils.FormatBytes(diskUsage)))

	// Graphique simple
	maxDisk := float64(1024 * 1024 * 1024) // 1GB max pour l'échelle
	if float64(diskUsage) > maxDisk {
		maxDisk = float64(diskUsage)
	}

	diskPercent := float64(diskUsage) / maxDisk * 100
	barWidth := 40
	filled := int(diskPercent / 100 * float64(barWidth))

	filledBar := lipgloss.NewStyle().Foreground(lipgloss.Color("#14b8a6")).Render(strings.Repeat("█", filled)) // teal pour disque
	emptyBar := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", barWidth-filled))
	doc.WriteString(fmt.Sprintf("[%s%s] %s\n", filledBar, emptyBar, utils.FormatBytes(diskUsage)))

	return doc.String()
}

func (m model) renderContainerEnv() string {
	doc := strings.Builder{}

	doc.WriteString(metricLabelStyle.Render("🔧 Environment Variables"))
	doc.WriteString("\n\n")

	if len(m.containerSingleView.EnvVars) == 0 {
		doc.WriteString("No environment variables found.")
		return doc.String()
	}

	// Trier les clés pour un affichage cohérent
	var keys []string
	for key := range m.containerSingleView.EnvVars {
		keys = append(keys, key)
	}

	// Trier les clés par ordre alphabétique
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Afficher les variables
	for _, key := range keys {
		value := m.containerSingleView.EnvVars[key]

		// Limiter la longueur de la valeur pour l'affichage
		if len(value) > 60 {
			value = value[:57] + "..."
		}

		line := fmt.Sprintf("%-20s = %s",
			metricLabelStyle.Render(key+":"),
			metricValueStyle.Render(value))
		doc.WriteString(line)
		doc.WriteString("\n")
	}

	return doc.String()
}
