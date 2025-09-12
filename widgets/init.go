package widgets

import (
	"fmt"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Initialisation du modèle
func InitialModel() Model {
	apk, err := app.NewDockerManager()
	if err != nil {
		panic(err)
	}
	return InitialModelWithManager(apk)
}

// InitialModelWithManager creates a model with an existing docker manager
func InitialModelWithManager(apk *app.DockerManager) Model {

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
		{Title: "Name", Width: 16},
		{Title: "Status", Width: 12},
		{Title: "Project", Width: 20},
		{Title: "Ports", Width: 20},
	}
	ct := table.New(
		table.WithColumns(ctColumns),
		table.WithFocused(true),
	)
	cs := table.DefaultStyles()
	cs.Header = cs.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	cs.Selected = cs.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	ct.SetStyles(cs)

	// Network table
	networkColumns := []table.Column{
		{Title: "Interface", Width: 12},
		{Title: "Status", Width: 8},
		{Title: "IP Addresses", Width: 25},
		{Title: "RX", Width: 12},
		{Title: "TX", Width: 12},
	}
	networkTable := table.New(
		table.WithColumns(networkColumns),
		table.WithFocused(false),
	)
	networkStyle := table.DefaultStyles()
	networkStyle.Header = networkStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	// networkStyle.Cell = networkStyle.Cell.Foreground(lipgloss.Color("255"))
	// networkTable.SetStyles(networkStyle)
	networkStyle.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	networkTable.SetStyles(networkStyle)

	searchInput := textinput.New()
	searchInput.Placeholder = "Search a process..."
	searchInput.Prompt = "/"
	searchInput.CharLimit = 50
	searchInput.Width = 30

	progOpts := []progress.Option{
		progress.WithWidth(progressBarWidth),
		progress.WithDefaultGradient(),
	}

	m := Model{
		Network: model.NetworkModel{
			NetworkTable: networkTable,
			Nav: networkNav,
			SelectedItem: model.ContainerTabGeneral,
		},
		Monitor: model.MonitorModel{
			ProcessTable:       t,
			Container:          ct,
			CpuProgress:        progress.New(progOpts...),
			MemProgress:        progress.New(progOpts...),
			SwapProgress:       progress.New(progOpts...),
			DiskProgress:       make(map[string]progress.Model),
			App:                apk,
			ContainerMenuState: ContainerMenuHidden,
			SelectedContainer:  nil,
			ContainerMenuItems: containerMenuItems,
			ContainerViewState: ContainerViewNone,
			ContainerTabs:      containerTabs,
			CpuHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			MemoryHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			NetworkRxHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			NetworkTxHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			DiskReadHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
			DiskWriteHistory: model.DataHistory{
				MaxPoints: 60,
				Points:    make([]model.DataPoint, 0),
			},
		},
		Ui: model.UIModel{
			Tabs:        menu,
			SelectedTab: 0,
			ActiveView:  -1,
			SearchInput: searchInput,
			SearchMode:  false,
			Viewport:    viewport.New(100, 20),
			MinWidth:    40,
			MinHeight:   10,
		},
		LastChartUpdate:   time.Now(),
		ScrollSensitivity: 3,
		MouseEnabled:      true,
		EnableAnimations:  true,
		ContainerTab:      model.ContainerTabGeneral,
		ProgressBars:      make(map[string]progress.Model),
	}

	m.Monitor.ProcessTable.Focus()
	m.updateContainerTable(containers)
	m.Monitor.Container.Focus()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		info.UpdateSystemInfo(),
		resource.UpdateCPUInfo(),
		resource.UpdateMemoryInfo(),
		resource.UpdateDiskInfo(),
		resource.UpdateNetworkInfo(),
		proc.UpdateProcesses(),
		m.Monitor.App.UpdateApp(),
	)
}
