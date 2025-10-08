package widgets

import (
	"fmt"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	info "github.com/System-Pulse/server-pulse/system/informations"
	"github.com/System-Pulse/server-pulse/system/logs"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/system/security"
	"github.com/System-Pulse/server-pulse/utils"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
		{Title: "Health", Width: 12},
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
		table.WithFocused(true),
	)
	networkStyle := table.DefaultStyles()
	networkStyle.Header = networkStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	networkStyle.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	networkTable.SetStyles(networkStyle)

	connectionsColumns := []table.Column{
		{Title: "Proto", Width: 6},
		{Title: "Recv-Q", Width: 8},
		{Title: "Send-Q", Width: 8},
		{Title: "Local Address", Width: 22},
		{Title: "Foreign Address", Width: 20},
		{Title: "State", Width: 12},
		{Title: "PID/Program", Width: 18},
	}
	connectionsTable := table.New(
		table.WithColumns(connectionsColumns),
		table.WithFocused(true),
		table.WithHeight(11),
	)
	connectionsTable.SetStyles(networkStyle)

	routesColumns := []table.Column{
		{Title: "Destination", Width: 18},
		{Title: "Gateway", Width: 12},
		{Title: "Genmask", Width: 16},
		{Title: "Flags", Width: 8},
		{Title: "Iface", Width: 16},
	}
	routesTable := table.New(
		table.WithColumns(routesColumns),
		table.WithFocused(true),
		table.WithHeight(11),
	)
	routesTable.SetStyles(networkStyle)

	dnsColumns := []table.Column{
		{Title: "", Width: 20},
	}
	dnsTable := table.New(
		table.WithColumns(dnsColumns),
		table.WithFocused(true),
		table.WithHeight(11),
	)
	dnsTable.SetStyles(networkStyle)

	diagnosticColumns := []table.Column{
		{Title: "Security Check", Width: 30},
		{Title: "Performances", Width: 12},
		{Title: "Logs", Width: 40},
	}

	diagnosticTable := table.New(
		table.WithColumns(diagnosticColumns),
		table.WithFocused(true),
	)
	diagnosticTable.SetStyles(networkStyle)

	searchInput := textinput.New()
	searchInput.Placeholder = "Search a process..."
	searchInput.Prompt = "/"
	searchInput.CharLimit = 50
	searchInput.Width = 30

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Enter root password..."
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 50
	passwordInput.Width = 30

	// Initialize authentication state
	isRoot := utils.IsRoot()
	sudoAvailable := isSudoAvailable()
	canRunSudo := canRunSudo()

	progOpts := []progress.Option{
		progress.WithWidth(v.ProgressBarWidth),
		progress.WithDefaultGradient(),
	}

	securityManager := security.NewSecurityManager()
	// Initialize SecurityManager with current auth state
	securityManager.IsRoot = isRoot
	securityManager.CanUseSudo = canRunSudo
	securityManager.SudoPassword = "" // No password initially

	securityColumns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 15},
		{Title: "Details", Width: 40},
	}
	securityTable := table.New(
		table.WithColumns(securityColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)

	securityTable.SetStyles(tableStyle)

	portsColumns := []table.Column{
		{Title: "Port", Width: 10},
		{Title: "Service", Width: 20},
		{Title: "Protocol", Width: 10},
		{Title: "PID", Width: 10},
	}

	portsTable := table.New(
		table.WithColumns(portsColumns),
		table.WithFocused(true),
	)

	portsTable.SetStyles(tableStyle)
	spinnerModel := getRandomSpinner()

	firewallColumns := []table.Column{
		{Title: "Firewall Rule", Width: 100},
	}

	firewallTable := table.New(
		table.WithColumns(firewallColumns),
		table.WithFocused(true),
	)

	firewallTable.SetStyles(tableStyle)

	autoBanColumns := []table.Column{
		{Title: "Jail/Service Details", Width: 100},
	}

	autoBanTable := table.New(
		table.WithColumns(autoBanColumns),
		table.WithFocused(true),
	)

	autoBanTable.SetStyles(tableStyle)

	// Logs table
	logsColumns := []table.Column{
		{Title: "Time", Width: 20},
		{Title: "Level", Width: 8},
		{Title: "Service", Width: 20},
		{Title: "Message", Width: 70},
	}

	logsTable := table.New(
		table.WithColumns(logsColumns),
		table.WithFocused(true),
		table.WithHeight(11),
	)

	logsTable.SetStyles(tableStyle)

	// Log filter inputs
	logSearchInput := textinput.New()
	logSearchInput.Placeholder = "Search logs..."
	logSearchInput.CharLimit = 100
	logSearchInput.Width = 30

	logServiceInput := textinput.New()
	logServiceInput.Placeholder = "Filter by service..."
	logServiceInput.CharLimit = 50
	logServiceInput.Width = 25

	logTimeRangeInput := textinput.New()
	logTimeRangeInput.Placeholder = "e.g., '2 hours ago', '2025-01-08', '3 days ago'"
	logTimeRangeInput.CharLimit = 50
	logTimeRangeInput.Width = 50

	// Initialize log manager
	logManager := logs.NewLogManager()

	// Default log filters
	defaultLogFilters := logs.LogFilters{
		TimeRange: "24h",
		Level:     logs.LogLevelAll,
		Limit:     100,
	}

	logsViewport := viewport.New(100, 20)
	m := Model{
		LogsViewport: logsViewport,
		HelpSystem:   NewHelpSystem(),
		Network: model.NetworkModel{
			NetworkTable:     networkTable,
			ConnectionsTable: connectionsTable,
			RoutesTable:      routesTable,
			DNSTable:         dnsTable,
			Nav:              v.NetworkNav,
			SelectedItem:     model.NetworkTabInterface,
			PingInput: func() textinput.Model {
				ti := textinput.New()
				ti.Placeholder = "Enter host or IP to ping (e.g., 8.8.8.8)"
				ti.CharLimit = 100
				ti.Width = 40
				return ti
			}(),
			TracerouteInput: func() textinput.Model {
				ti := textinput.New()
				ti.Placeholder = "Enter host or IP to trace (e.g., google.com)"
				ti.CharLimit = 100
				ti.Width = 40
				return ti
			}(),
			ConnectivityMode: model.ConnectivityModeNone,
			AuthState:        model.AuthNotRequired,
			AuthMessage:      "",
			IsRoot:           isRoot,
			SudoAvailable:    sudoAvailable,
			CanRunSudo:       canRunSudo,
		},
		Diagnostic: model.DiagnosticModel{
			DiagnosticTable:     diagnosticTable,
			Nav:                 v.DiagnosticNav,
			SelectedItem:        model.DiagnosticSecurityChecks,
			SecurityManager:     securityManager,
			SecurityTable:       securityTable,
			PortsTable:          portsTable,
			FirewallTable:       firewallTable,
			AutoBanTable:        autoBanTable,
			LogsTable:           logsTable,
			LogManager:          logManager,
			LogFilters:          defaultLogFilters,
			LogSearchInput:      logSearchInput,
			LogServiceInput:     logServiceInput,
			LogTimeRangeInput:   logTimeRangeInput,
			LogLevelSelected:    0,
			LogTimeSelected:     3, // Default to "24h" (index: All=0, 5m=1, 1h=2, 24h=3, 7d=4, Custom=5)
			CustomTimeInputMode: false,
			DomainInput: func() textinput.Model {
				ti := textinput.New()
				ti.Placeholder = "Enter domain name (e.g., google.com)"
				ti.CharLimit = 100
				ti.Width = 40
				return ti
			}(),
			DomainInputMode: false,
			Password:        passwordInput,
			AuthState:       model.AuthNotRequired,
			AuthMessage:     "",
			IsRoot:          isRoot,
			SudoAvailable:   sudoAvailable,
			CanRunSudo:      canRunSudo,
			AuthTimer:       0,
		},
		Monitor: model.MonitorModel{
			ProcessTable:       t,
			Container:          ct,
			CpuProgress:        progress.New(progOpts...),
			MemProgress:        progress.New(progOpts...),
			SwapProgress:       progress.New(progOpts...),
			DiskProgress:       make(map[string]progress.Model),
			App:                apk,
			ContainerMenuState: v.ContainerMenuHidden,
			SelectedContainer:  nil,
			ContainerMenuItems: v.ContainerMenuItems,
			ContainerViewState: v.ContainerViewNone,
			ContainerTabs:      v.ContainerTabs,
			ContainerLogsPagination: model.ContainerLogsPagination{
				PageSize:    100,
				CurrentPage: 1,
				TotalPages:  1,
				Lines:       []string{},
			},
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
			ContainerHistories: make(map[string]model.ContainerHistory),
		},
		Ui: model.UIModel{
			State:         model.StateHome,
			Tabs:          v.Menu,
			SelectedTab:   0,
			ActiveView:    -1,
			SearchInput:   searchInput,
			SearchMode:    false,
			Viewport:      viewport.New(100, 20),
			ContentHeight: 20,
			Spinner:       spinnerModel,
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
