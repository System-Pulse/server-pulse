package widgets

import (
	"fmt"

	"github.com/System-Pulse/server-pulse/system/app"
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
	apk, err := app.NewDockerManager()
	if err != nil {
		panic(err)
	}

	containers, err := apk.RefreshContainers()
	if err != nil {
		// Gérer l'erreur ou logger
		fmt.Printf("Erreur lors du chargement des conteneurs: %v\n", err)
	}

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
	// containers table
	ctColumns := []table.Column{
		{Title: "ID", Width: 12},
		{Title: "Image", Width: 12},
		{Title: "Name", Width: 12},
		{Title: "Status", Width: 12},
		{Title: "Project", Width: 20},
	}
	ct := table.New(
		table.WithColumns(ctColumns),
		table.WithFocused(true),
	)
	cs := table.DefaultStyles()
	cs.Header = cs.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	cs.Selected = cs.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	ct.SetStyles(cs)

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
	m := model{
		tabs:         menu,
		selectedTab:  0,
		activeView:   -1,
		processTable: t,
		container:    ct,
		searchInput:  searchInput,
		searchMode:   false,
		cpuProgress:  progress.New(progOpts...),
		memProgress:  progress.New(progOpts...),
		swapProgress: progress.New(progOpts...),
		diskProgress: make(map[string]progress.Model),
		viewport:     viewport.New(100, 20),
		app:          apk,
	}

	m.updateContainerTable(containers)

	return m
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
		m.app.UpdateApp(),
	)
}
