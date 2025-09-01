package widgets

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/utils"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Initialisation du modèle
func InitialModel() model {
	columns := []table.Column{
		{Title: "PID", Width: 8},
		{Title: "User", Width: 12},
		{Title: "CPU%", Width: 8},
		{Title: "Mem%", Width: 8},
		{Title: "Command", Width: 30},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	t.SetStyles(s)

	searchInput := textinput.New()
	searchInput.Placeholder = "Rechercher un processus..."
	searchInput.Prompt = "/"
	searchInput.CharLimit = 50
	searchInput.Width = 30

	progOpts := []progress.Option{
		progress.WithWidth(progressBarWidth),
		progress.WithDefaultGradient(),
	}
	dashboard := []string{"Monitor", "Diagnostic", "Network", "Reporting"}
	monitor := []string{"System", "Process", "Application"}
	menu := Menu{
		DashBoard: dashboard,
		Monitor:   monitor,
	}
	return model{
		tabs:         menu,
		selectedTab:  0,
		activeView:   -1,
		processTable: t,
		searchInput:  searchInput,
		searchMode:   false,      
		cpuProgress:  progress.New(progOpts...),
		memProgress:  progress.New(progOpts...),
		swapProgress: progress.New(progOpts...),
		diskProgress: make(map[string]progress.Model),
		viewport:     viewport.New(100, 20),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		info.UpdateSystemInfo(),
		resource.UpdateCPUInfo(),
		resource.UpdateMemoryInfo(),
		resource.UpdateDiskInfo(),
		resource.UpdateNetworkInfo(),
		proc.UpdateProcesses(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.renderHome()) + 2
		footerHeight := lipgloss.Height(m.renderFooter())
		verticalMargin := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargin
		m.processTable.SetWidth(msg.Width)
		m.processTable.SetHeight(m.viewport.Height - 1)
		progWidth := min(max(msg.Width/3, 20), progressBarWidth)
		m.cpuProgress.Width = progWidth
		m.memProgress.Width = progWidth
		m.swapProgress.Width = progWidth
		for k, p := range m.diskProgress {
			p.Width = progWidth
			m.diskProgress[k] = p
		}

		if !m.ready {
			m.ready = true
		}
		return m, nil
	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "esc", "enter":
				m.searchMode = false
				m.processTable.Focus()
				return m, m.updateProcessTable()
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Scrolling/navigation
		if m.isMonitorActive && m.selectedMonitor == 1 {
			switch msg.String() {
			case "up", "k":
				m.processTable.MoveUp(1)
			case "down", "j":
				m.processTable.MoveDown(1)
			case "pageup":
				m.processTable.MoveUp(10)
			case "pagedown":
				m.processTable.MoveDown(10)
			case "home":
				m.processTable.GotoTop()
			case "end":
				m.processTable.GotoBottom()
			case "/":
				m.searchMode = true
				m.searchInput.Focus()
				return m, nil
			}
		} else if m.activeView != -1 {
			switch msg.String() {
			case "up", "k":
				m.viewport.ScrollUp(1)
			case "down", "j":
				m.viewport.ScrollDown(1)
			}
		}
		
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.activeView == -1 {
				m.activeView = m.selectedTab
				if m.activeView == 0 { // Monitor
					m.isMonitorActive = true
				}
			} else if m.isMonitorActive {
				// NOTHING
			}
		case "b", "esc":
			if m.searchMode {
				m.searchMode = false
				m.processTable.Focus()
			} else if m.isMonitorActive {
				m.isMonitorActive = false
				m.activeView = -1
			} else if m.activeView != -1 {
				m.activeView = -1
			}
		case "tab", "right", "l":
			if m.activeView == -1 {
				m.selectedTab = (m.selectedTab + 1) % len(m.tabs.DashBoard)
			} else if m.isMonitorActive {
				m.selectedMonitor = (m.selectedMonitor + 1) % len(m.tabs.Monitor)
			}
		case "shift+tab", "left", "h":
			if m.activeView == -1 {
				m.selectedTab = (m.selectedTab - 1 + len(m.tabs.DashBoard)) % len(m.tabs.DashBoard)
			} else if m.isMonitorActive {
				m.selectedMonitor = (m.selectedMonitor - 1 + len(m.tabs.Monitor)) % len(m.tabs.Monitor)
			}
		case "1":
			if m.activeView == -1 {
				m.selectedTab = 0
			}
		case "2":
			if m.activeView == -1 {
				m.selectedTab = 1
			}
		case "3":
			if m.activeView == -1 {
				m.selectedTab = 2
			}
		case "4":
			if m.activeView == -1 {
				m.selectedTab = 3
			}
		case "k":
			if m.isMonitorActive && m.selectedMonitor == 1 && len(m.processTable.SelectedRow()) > 0 {
				selectedPID := m.processTable.SelectedRow()[0]
				pid, _ := strconv.Atoi(selectedPID)
				process, _ := os.FindProcess(pid)
				if process != nil {
					_ = process.Kill()
				}
				return m, proc.UpdateProcesses()
			}
		case "s":
			if m.isMonitorActive && m.selectedMonitor == 1 {
				sort.Slice(m.processes, func(i, j int) bool {
					return m.processes[i].CPU > m.processes[j].CPU
				})
				return m, m.updateProcessTable()
			}
		case "m":
			if m.isMonitorActive && m.selectedMonitor == 1 {
				sort.Slice(m.processes, func(i, j int) bool {
					return m.processes[i].Mem > m.processes[j].Mem
				})
				return m, m.updateProcessTable()
			}
		}

	case info.SystemMsg:
		m.system = info.SystemInfo(msg)
	case resource.CpuMsg:
		m.cpu = resource.CPUInfo(msg)
		cmds = append(cmds, m.cpuProgress.SetPercent(m.cpu.Usage/100))
	case resource.MemoryMsg:
		m.memory = resource.MemoryInfo(msg)
		cmds = append(cmds, m.memProgress.SetPercent(m.memory.Usage/100))
		cmds = append(cmds, m.swapProgress.SetPercent(m.memory.SwapUsage/100))
	case resource.DiskMsg:
		m.disks = []resource.DiskInfo(msg)
		for _, disk := range m.disks {
			if _, ok := m.diskProgress[disk.Mountpoint]; !ok && disk.Total > 0 {
				progOpts := []progress.Option{
					progress.WithWidth(m.cpuProgress.Width),
					progress.WithDefaultGradient(),
				}
				m.diskProgress[disk.Mountpoint] = progress.New(progOpts...)
			}
			if disk.Total > 0 {
				prog := m.diskProgress[disk.Mountpoint]
				cmds = append(cmds, prog.SetPercent(disk.Usage/100))
				m.diskProgress[disk.Mountpoint] = prog
			}
		}
	case resource.NetworkMsg:
		m.network = resource.NetworkInfo(msg)
	case proc.ProcessMsg:
		m.processes = []proc.ProcessInfo(msg)
		return m, m.updateProcessTable()
	case utils.ErrMsg:
		m.err = msg
	case utils.TickMsg:
		cmds = append(cmds,
			tick(),
			info.UpdateSystemInfo(),
			resource.UpdateCPUInfo(),
			resource.UpdateMemoryInfo(),
			resource.UpdateDiskInfo(),
			resource.UpdateNetworkInfo(),
			proc.UpdateProcesses(),
		)
	case progress.FrameMsg:
		var progCmd tea.Cmd
		var updatedModel tea.Model

		updatedModel, progCmd = m.cpuProgress.Update(msg)
		m.cpuProgress = updatedModel.(progress.Model)
		cmds = append(cmds, progCmd)

		updatedModel, progCmd = m.memProgress.Update(msg)
		m.memProgress = updatedModel.(progress.Model)
		cmds = append(cmds, progCmd)

		updatedModel, progCmd = m.swapProgress.Update(msg)
		m.swapProgress = updatedModel.(progress.Model)
		cmds = append(cmds, progCmd)

		for key, p := range m.diskProgress {
			updatedModel, progCmd := (p).Update(msg)

			newModel := updatedModel.(progress.Model)
			m.diskProgress[key] = newModel

			cmds = append(cmds, progCmd)
		}
	}

	if m.isMonitorActive && m.selectedMonitor == 1 {
		if !m.searchMode {
			m.processTable, cmd = m.processTable.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else if m.activeView != -1 {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.err != nil {
		return fmt.Sprintf("Erreur: %v\n", m.err)
	}

	var currentView string
	
	if m.activeView != -1 {
		switch m.activeView {
		case 0: // Monitor
			if m.isMonitorActive {
				currentView = m.renderMonitor()
			} else {
				currentView = m.renderSystem()
			}
		case 1: // Diagnostic
			currentView = m.renderDignostics()
		case 2: // Network
			currentView = m.renderNetwork()
		case 3: // Reporting
			currentView = m.renderReporting()
		}
	} else {
		currentView = m.renderSystem()
	}

	home := m.renderHome()
	tabs := m.renderTabs()
	footer := m.renderFooter()

	var mainContent string
	if m.isMonitorActive && m.selectedMonitor == 1 {
		if m.searchMode {
			searchBar := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("57")).
				Padding(0, 1).
				MarginBottom(1).
				Render(m.searchInput.View())
			mainContent = lipgloss.JoinVertical(lipgloss.Left, searchBar, m.processTable.View())
		} else {
			mainContent = m.processTable.View()
		}
	} else {
		if m.activeView != -1 {
			m.viewport.SetContent(currentView)
		}
		mainContent = m.viewport.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		home,
		tabs,
		mainContent,
		footer,
	)
}

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
	return "APPLICATION"
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

func (m *model) updateProcessTable() tea.Cmd {
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

func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return utils.TickMsg(t)
	})
}
