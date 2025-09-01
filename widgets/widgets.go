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
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Initialisation du modèle
func InitialModel() model {
	// Configuration de la table des processus
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
	progOpts := []progress.Option{
		progress.WithWidth(progressBarWidth),
		progress.WithDefaultGradient(),
	}
	dashboard := []string{"Dashboard", "Processes", "Network"}
	monitor := []string{"System", "Application"}
	menu := Menu{
		DashBoard: dashboard,
		Monitor: monitor,
	}
	return model{
		tabs:         menu,
		selectedTab:  0,
		processTable: t,
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
		headerHeight := lipgloss.Height(m.renderTabs()) + 2
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
		if m.selectedTab != 1 {
			switch msg.String() {
			case "up", "u":
				m.viewport.ScrollUp(1)
			case "down", "j":
				m.viewport.ScrollDown(1)
			}
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			m.selectedTab = (m.selectedTab + 1) % len(m.tabs.DashBoard)
		case "shift+tab", "left", "h":
			m.selectedTab = (m.selectedTab - 1 + len(m.tabs.DashBoard)) % len(m.tabs.DashBoard)
		case "1":
			m.selectedTab = 0
		case "2":
			m.selectedTab = 1
		case "3":
			m.selectedTab = 2
		case "k":
			if m.selectedTab == 1 && len(m.processTable.SelectedRow()) > 0 {
				selectedPID := m.processTable.SelectedRow()[0]
				pid, _ := strconv.Atoi(selectedPID)
				process, _ := os.FindProcess(pid)
				if process != nil {
					_ = process.Kill()
				}
				return m, proc.UpdateProcesses()
			}
		case "s":
			sort.Slice(m.processes, func(i, j int) bool {
				return m.processes[i].CPU > m.processes[j].CPU
			})
			return m, m.updateProcessTable()
		case "m":
			sort.Slice(m.processes, func(i, j int) bool {
				return m.processes[i].Mem > m.processes[j].Mem
			})
			return m, m.updateProcessTable()
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

	if m.selectedTab == 1 {
		m.processTable, cmd = m.processTable.Update(msg)
	} else {
		var content string
		switch m.selectedTab {
		case 0:
			content = m.renderDashboard()
		case 2:
			content = m.renderNetwork()
		}
		m.viewport.SetContent(content)
		m.viewport, cmd = m.viewport.Update(msg)
	}
	cmds = append(cmds, cmd)

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
	switch m.selectedTab {
	case 0:
		currentView = m.renderDashboard()
	case 1:
		currentView = m.renderProcesses()
	case 2:
		currentView = m.renderNetwork()
	}

	if m.selectedTab != 1 {
		m.viewport.SetContent(currentView)
	}

	tabs := m.renderTabs()
	footer := m.renderFooter()

	var mainContent string
	if m.selectedTab == 1 {
		mainContent = m.processTable.View()
	} else {
		mainContent = m.viewport.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		tabs,
		mainContent,
		footer,
	)
}

func (m model) renderTabs() string {
	var tabs []string
	for i, t := range m.tabs.DashBoard {
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == m.selectedTab {
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

func (m model) renderDashboard() string {
	doc := strings.Builder{}
	systemInfo := fmt.Sprintf("Host: %s | OS: %s | Kernel: %s | Uptime: %s", m.system.Hostname, m.system.OS, m.system.Kernel, utils.FormatUptime(m.system.Uptime))
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("System Info"))
	doc.WriteString("\n")
	doc.WriteString(systemInfo)
	doc.WriteString("\n\n")
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

	content.WriteString(metricLabelStyle.Render("⚡ Active Processes"))
	content.WriteString("\n\n")
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
	switch m.selectedTab {
	case 0:
		footer += "[q] Quit • [tab/←→] Switch • [1-3] Direct tab • [r] Refresh"
	case 1:
		footer += "[q] Quit • [↑↓] Navigate • [k] Kill • [s] Sort CPU • [m] Sort Mem"
	case 2:
		footer += "[q] Quit • [tab/←→] Switch • [r] Refresh"
	default:
		footer += "[q] Quit • [tab/←→] Switch"
	}
	return footer
}

func (m *model) updateProcessTable() tea.Cmd {
	var rows []table.Row
	for _, p := range m.processes {
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
