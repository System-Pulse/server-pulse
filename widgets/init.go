package widgets

import (
	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"
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
