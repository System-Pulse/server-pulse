package widgets

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	system "github.com/System-Pulse/server-pulse/system/app"
	info "github.com/System-Pulse/server-pulse/system/informations"
	logs "github.com/System-Pulse/server-pulse/system/logs"
	"github.com/System-Pulse/server-pulse/system/performance"
	proc "github.com/System-Pulse/server-pulse/system/process"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/System-Pulse/server-pulse/system/network"
	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/system/security"
	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// ------------------------- Handlers for system-related and resource messages -------------------------

func (m Model) handleResourceAndProcessMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case info.SystemMsg:
		m.Monitor.System = info.SystemInfo(msg)
	case resource.CpuMsg:
		m.Monitor.Cpu = resource.CPUInfo(msg)
		cmds = append(cmds, m.Monitor.CpuProgress.SetPercent(m.Monitor.Cpu.Usage/100))
	case resource.MemoryMsg:
		m.Monitor.Memory = resource.MemoryInfo(msg)
		cmds = append(cmds, m.Monitor.MemProgress.SetPercent(m.Monitor.Memory.Usage/100))
		cmds = append(cmds, m.Monitor.SwapProgress.SetPercent(m.Monitor.Memory.SwapUsage/100))
	case resource.DiskMsg:
		m.Monitor.Disks = []resource.DiskInfo(msg)
		for _, disk := range m.Monitor.Disks {
			if _, ok := m.Monitor.DiskProgress[disk.Mountpoint]; !ok && disk.Total > 0 {
				progOpts := []progress.Option{
					progress.WithWidth(m.Monitor.CpuProgress.Width),
					progress.WithDefaultGradient(),
				}
				m.Monitor.DiskProgress[disk.Mountpoint] = progress.New(progOpts...)
			}
			if disk.Total > 0 {
				prog := m.Monitor.DiskProgress[disk.Mountpoint]
				cmds = append(cmds, prog.SetPercent(disk.Usage/100))
				m.Monitor.DiskProgress[disk.Mountpoint] = prog
			}
		}
	case resource.NetworkMsg:
		m.Network.NetworkResource = resource.NetworkInfo(msg)
		cmds = append(cmds, m.updateNetworkTable())
	case proc.ProcessMsg:
		m.Monitor.Processes = []proc.ProcessInfo(msg)
		return m, m.updateProcessTable()
	case performance.HealthMetricsMsg:
		if msg.Metrics != nil {
			m.Diagnostic.Performance.HealthMetrics = &model.HealthMetrics{
				IOWait:          msg.Metrics.IOWait,
				ContextSwitches: msg.Metrics.ContextSwitches,
				Interrupts:      msg.Metrics.Interrupts,
				StealTime:       msg.Metrics.StealTime,
				MajorFaults:     msg.Metrics.MajorFaults,
				MinorFaults:     msg.Metrics.MinorFaults,
			}
		}
		if msg.Score != nil {
			m.Diagnostic.Performance.HealthScore = &model.HealthScore{
				Score:           msg.Score.Score,
				Issues:          msg.Score.Issues,
				Recommendations: msg.Score.Recommendations,
				ChecksPerformed: msg.Score.ChecksPerformed,
			}
		}
		m.Diagnostic.Performance.HealthLoading = false
		return m, nil
	case performance.IOMetricsMsg:
		if msg.Error != nil {
			m.Diagnostic.Performance.IOLoading = false
			return m, nil
		}
		m.Diagnostic.Performance.IOMetrics = msg.Metrics
		m.Diagnostic.Performance.IOLoading = false
		return m, nil
	case performance.CPUMetricsMsg:
		if msg.Error != nil {
			m.Diagnostic.Performance.CPULoading = false
			return m, nil
		}
		m.Diagnostic.Performance.CPUMetrics = msg.Metrics
		m.Diagnostic.Performance.CPULoading = false
		return m, nil
	case performance.MemoryMetricsMsg:
		if msg.Error != nil {
			m.Diagnostic.Performance.MemoryLoading = false
			return m, nil
		}
		m.Diagnostic.Performance.MemoryMetrics = msg.Metrics
		m.Diagnostic.Performance.MemoryLoading = false
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

// ------------------------- Handlers for container-related messages -------------------------

func (m Model) handleContainerRelatedMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case system.ContainerMsg:
		containers := []system.Container(msg)
		if m.Monitor.SelectedContainer != nil && m.Monitor.ContainerLogsStreaming {
			for _, container := range containers {
				if container.ID == m.Monitor.SelectedContainer.ID {
					if strings.ToLower(container.Status) != "up" && m.Monitor.ContainerLogsStreaming {
						m.cleanupLogsStream()
						m.LastOperationMsg = fmt.Sprintf("Container stopped, streaming disabled (status: %s)", container.Status)
					}
					break
				}
			}
		}
		return m, m.updateContainerTable(containers)
	case system.ContainerDetailsMsg:
		details := system.ContainerDetails(msg)
		m.Monitor.ContainerDetails = &details
	case system.ContainerLogsMsg:
		logsMsg := system.ContainerLogsMsg(msg)
		m.Monitor.ContainerLogsLoading = false
		if logsMsg.Error != nil {
			if strings.Contains(logsMsg.Error.Error(), "streaming unavailable") {
				m.Monitor.ContainerLogs = fmt.Sprintf("Streaming not available: %v\nShowing static logs instead...", logsMsg.Error)
				return m, m.Monitor.App.GetContainerLogsCmd(m.Monitor.SelectedContainer.ID)
			} else {
				m.Monitor.ContainerLogs = fmt.Sprintf("Error loading logs: %v", logsMsg.Error)
			}
			m.Monitor.ContainerLogsPagination.Lines = []string{m.Monitor.ContainerLogs}
			m.Monitor.ContainerLogsPagination.TotalPages = 1
			m.Monitor.ContainerLogsPagination.CurrentPage = 1
		} else {
			m.Monitor.ContainerLogs = logsMsg.Logs
			m.setContainerLogs(m.Monitor.ContainerLogs)
		}
	case system.ContainerOperationMsg:
		opMsg := system.ContainerOperationMsg(msg)
		m.OperationInProgress = false
		m.LastOperationMsg = utils.FormatOperationMessage(opMsg.Operation, opMsg.Success, opMsg.Error)

		var refreshCmd tea.Cmd
		if opMsg.Success {
			refreshCmd = m.Monitor.App.UpdateApp()
		}
		return m, tea.Batch(refreshCmd, clearOperationMessage())
	case system.ExecShellMsg:
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: msg.ContainerID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case system.ContainerStatsChanMsg:
		statsMsg := system.ContainerStatsChanMsg(msg)
		go m.handleRealTimeStats(statsMsg.ContainerID, statsMsg.StatsChan)
		return m, nil
	}
	return m, nil
}

// ------------------------- Window size -------------------------

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Ui.Width = msg.Width
	m.Ui.Height = msg.Height

	if msg.Width < 60 || msg.Height < 15 {
		m.Ui.Ready = false
		return m, nil
	}
	m.Ui.Ready = true

	// Set help system width to match window width
	m.HelpSystem.SetWidth(msg.Width)
	m.updateViewportDimensions()

	m.LogsViewport.Width = msg.Width
	m.LogsViewport.Height = m.Ui.ContentHeight

	tableHeight := max(1, m.Ui.ContentHeight-3)

	m.Monitor.ProcessTable.SetWidth(msg.Width)
	m.Monitor.ProcessTable.SetHeight(tableHeight)

	m.Monitor.Container.SetWidth(msg.Width)
	m.Monitor.Container.SetHeight(tableHeight)

	m.Network.NetworkTable.SetWidth(msg.Width)
	m.Network.NetworkTable.SetHeight(tableHeight)

	progWidth := min(max(msg.Width/3, 20), v.ProgressBarWidth)
	m.Monitor.CpuProgress.Width = progWidth
	m.Monitor.MemProgress.Width = progWidth
	m.Monitor.SwapProgress.Width = progWidth

	for k, p := range m.Monitor.DiskProgress {
		p.Width = progWidth
		m.Monitor.DiskProgress[k] = p
	}

	return m, nil
}

// ------------------------- Key handling -------------------------

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ConfirmationVisible {
		return m.handleConfirmationKeys(msg)
	}
	if m.Monitor.ContainerMenuState == model.ContainerMenuState(1) { // ContainerMenuVisible
		return m.handleContainerMenuKeys(msg)
	}
	if m.Ui.SearchMode {
		return m.handleSearchKeys(msg)
	}

	switch m.Ui.State {
	case model.StateHome:
		return m.handleHomeKeys(msg)
	case model.StateMonitor:
		return m.handleMonitorKeys(msg)
	case model.StateSystem:
		return m.handleSystemKeys(msg)
	case model.StateProcess:
		return m.handleProcessKeys(msg)
	case model.StateContainers:
		return m.handleContainersKeys(msg)
	case model.StateContainer:
		return m.handleContainerSingleKeys(msg)
	case model.StateContainerLogs:
		return m.handleContainerLogsKeys(msg)
	case model.StateNetwork:
		return m.handleNetworkKeys(msg)
	case model.StateDiagnostics:
		return m.handleDiagnosticsKeys(msg)
	case model.StateCertificateDetails:
		return m.handleCertificateDetailsKeys(msg)
	case model.StateSSHRootDetails:
		return m.handleSSHRootDetailsKeys(msg)
	case model.StateOpenedPortsDetails:
		return m.handleOpenedPortsDetailsKeys(msg)
	case model.StateFirewallDetails:
		return m.handleFirewallDetailsKeys(msg)
	case model.StateAutoBanDetails:
		return m.handleAutoBanDetailsKeys(msg)
	case model.StateReporting:
		return m.handleReportingKeys(msg)
	case model.StatePerformance, model.StateInputOutput, model.StateSystemHealth, model.StateCPU, model.StateMemory, model.StateQuickTests:
		return m.handlePerformanceKeys(msg)
	}

	return m, nil
}

func (m Model) handleHomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "1", "2", "3", "4":
		tabIndex, _ := strconv.Atoi(msg.String())
		m.Ui.SelectedTab = tabIndex - 1
	case "tab", "right", "l":
		m.Ui.SelectedTab = (m.Ui.SelectedTab + 1) % len(m.Ui.Tabs.DashBoard)
	case "shift+tab", "left", "h":
		m.Ui.SelectedTab = (m.Ui.SelectedTab - 1 + len(m.Ui.Tabs.DashBoard)) % len(m.Ui.Tabs.DashBoard)
	case "enter":
		switch m.Ui.SelectedTab {
		case 0:
			m.setState(model.StateMonitor)
			m.Ui.ActiveView = m.Ui.SelectedTab
		case 1:
			m.setState(model.StateDiagnostics)
			m.Ui.ActiveView = m.Ui.SelectedTab
			// Auto-load security checks if we're on security tab and they're not loaded
			if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
				domain := m.Diagnostic.DomainInput.Value()
				return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
			}
		case 2:
			m.setState(model.StateNetwork)
			m.Ui.ActiveView = m.Ui.SelectedTab
		case 3:
			m.setState(model.StateReporting)
			m.Ui.ActiveView = m.Ui.SelectedTab
		}
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleGeneralKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "1", "2", "3":
		monitorIndex, _ := strconv.Atoi(msg.String())
		m.Ui.SelectedMonitor = monitorIndex - 1

		switch m.Ui.SelectedMonitor {
		case 0:
			m.setState(model.StateSystem)
		case 1:
			m.setState(model.StateProcess)
		case 2:
			m.setState(model.StateContainers)
		}
	case "tab", "right", "l":
		m.Ui.SelectedMonitor = (m.Ui.SelectedMonitor + 1) % len(m.Ui.Tabs.Monitor)
		switch m.Ui.SelectedMonitor {
		case 0:
			m.setState(model.StateSystem)
		case 1:
			m.setState(model.StateProcess)
		case 2:
			m.setState(model.StateContainers)
		}
	case "shift+tab", "left", "h":
		m.Ui.SelectedMonitor = (m.Ui.SelectedMonitor - 1 + len(m.Ui.Tabs.Monitor)) % len(m.Ui.Tabs.Monitor)
		switch m.Ui.SelectedMonitor {
		case 0:
			m.setState(model.StateSystem)
		case 1:
			m.setState(model.StateProcess)
		case 2:
			m.setState(model.StateContainers)
		}
	case "b", "esc":
		m.goBack()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleMonitorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleGeneralKeys(msg)
}

// ------------------------- Handlers spécifiques à chaque état -------------------------

func (m Model) handleSystemKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "up", "k":
		m.Ui.Viewport.ScrollUp(1)
	case "down", "j":
		m.Ui.Viewport.ScrollDown(1)
	default:
		return m.handleGeneralKeys(msg)
	}
	return m, nil
}

func (m Model) handleProcessKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "/":
		m.Ui.SearchMode = true
		m.Ui.SearchInput.Focus()
	case "up", "u":
		m.Monitor.ProcessTable.MoveUp(1)
	case "down", "j":
		m.Monitor.ProcessTable.MoveDown(1)
	case "s":
		sort.Slice(m.Monitor.Processes, func(i, j int) bool { return m.Monitor.Processes[i].CPU > m.Monitor.Processes[j].CPU })
		return m, m.updateProcessTable()
	case "m":
		sort.Slice(m.Monitor.Processes, func(i, j int) bool { return m.Monitor.Processes[i].Mem > m.Monitor.Processes[j].Mem })
		return m, m.updateProcessTable()
	case "k": // stop process
		if len(m.Monitor.ProcessTable.SelectedRow()) > 0 {
			pid, _ := strconv.Atoi(m.Monitor.ProcessTable.SelectedRow()[0])
			if err := proc.StopProcess(pid); err == nil {
			}
			return m, proc.UpdateProcesses()
		}
	default:
		return m.handleGeneralKeys(msg)
	}
	return m, nil
}

func (m Model) handleContainersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "/":
		m.Ui.SearchMode = true
		m.Ui.SearchInput.Focus()
	case "up", "k":
		m.Monitor.Container.MoveUp(1)
	case "down", "j":
		m.Monitor.Container.MoveDown(1)
	case "enter":
		if len(m.Monitor.Container.SelectedRow()) > 0 {
			selectedRow := m.Monitor.Container.SelectedRow()
			containerID := selectedRow[0]
			containers, _ := m.Monitor.App.RefreshContainers()
			found := false
			for _, container := range containers {
				if container.ID == containerID {
					m.Monitor.SelectedContainer = &container
					found = true
					break
				}
			}
			if !found {
				m.Monitor.SelectedContainer = &system.Container{ID: containerID, Name: selectedRow[2]}
			}
			m.Monitor.ContainerMenuState = model.ContainerMenuState(1) // ContainerMenuVisible
			m.Monitor.SelectedMenuItem = 0
		}
	default:
		return m.handleGeneralKeys(msg)
	}
	return m, nil
}

func (m Model) handleContainerSingleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "b", "esc":
		m.Monitor.SelectedContainer = nil
		m.goBack()
		return m, nil
	case "tab", "right", "l":
		m.ContainerTab = model.ContainerTab((int(m.ContainerTab) + 1) % len(m.Monitor.ContainerTabs))
		return m, nil
	case "shift+tab", "left", "h":
		newTab := int(m.ContainerTab) - 1
		if newTab < 0 {
			newTab = len(m.Monitor.ContainerTabs) - 1
		}
		m.ContainerTab = model.ContainerTab(newTab)
		return m, nil
	case "1":
		m.ContainerTab = model.ContainerTabGeneral
		return m, nil
	case "2":
		m.ContainerTab = model.ContainerTabCPU
		return m, nil
	case "3":
		m.ContainerTab = model.ContainerTabMemory
		return m, nil
	case "4":
		m.ContainerTab = model.ContainerTabNetwork
		return m, nil
	// case "5":
	// 	m.ContainerTab = model.ContainerTabDisk
	// 	return m, nil
	case "5":
		m.ContainerTab = model.ContainerTabEnv
		return m, nil
	case "r":
		if m.Monitor.SelectedContainer != nil {
			return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
		}
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleContainerLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "b", "esc":
		m.Monitor.ContainerLogs = ""
		m.cleanupLogsStream() // 🔥
		m.goBack()
	case "up", "k":
		if m.Monitor.ContainerLogsPagination.CurrentPage > 1 {
			m.Monitor.ContainerLogsPagination.CurrentPage--
			m.updateLogsViewport()
		}
		return m, nil
	case "down", "j":
		if m.Monitor.ContainerLogsPagination.CurrentPage < m.Monitor.ContainerLogsPagination.TotalPages {
			m.Monitor.ContainerLogsPagination.CurrentPage++
			m.updateLogsViewport()
		}
		return m, nil
	case "s": // toggle streaming
		if m.Monitor.ContainerLogsStreaming {
			m.cleanupLogsStream()
			m.Monitor.ContainerLogsLoading = true
			return m, m.Monitor.App.GetContainerLogsCmd(m.Monitor.SelectedContainer.ID)
		} else {
			m.Monitor.ContainerLogsLoading = true
			m.Monitor.ContainerLogsPagination.Clear()
			return m, m.Monitor.App.StartLogsStreamCmd(m.Monitor.SelectedContainer.ID)
		}
	case "r": // refresh
		if m.Monitor.ContainerLogsStreaming {
			m.cleanupLogsStream()
		}
		m.Monitor.ContainerLogsLoading = true
		m.Monitor.ContainerLogsPagination.Clear()
		return m, m.Monitor.App.GetContainerLogsCmd(m.Monitor.SelectedContainer.ID)
	case "pageup":
		m.Ui.Viewport.PageUp()
		return m, nil
	case "pagedown":
		m.Ui.Viewport.PageDown()
		return m, nil
	case "home":
		m.Monitor.ContainerLogsPagination.CurrentPage = 1
		m.updateLogsViewport()
		return m, nil
	case "end":
		m.Monitor.ContainerLogsPagination.CurrentPage = m.Monitor.ContainerLogsPagination.TotalPages
		m.updateLogsViewport()
		return m, nil
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleNetworkKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Network.AuthState == model.AuthRequired {
		switch msg.String() {
		case "enter":
			if m.Diagnostic.Password.Value() != "" {
				m.Network.AuthState = model.AuthInProgress
				m.Network.AuthMessage = "Authenticating..."
				// Try to authenticate
				err := m.setRoot()
				if err != nil {
					m.Network.AuthState = model.AuthFailed
					m.Network.AuthMessage = fmt.Sprintf("Authentication failed: %v", err)
					if !m.SudoAvailable {
						m.Network.AuthMessage += "\nSudo is not available. Please run as root."
					}
					m.Diagnostic.Password.Reset()
				} else {
					m.Network.AuthState = model.AuthSuccess
					m.Network.AuthMessage = "Authentication successful!"
					m.AsRoot = true
					m.Network.AuthTimer = 2

					m.Diagnostic.SecurityManager.IsRoot = false // User authenticated via sudo, not actual root
					m.Diagnostic.SecurityManager.CanUseSudo = true
					m.Diagnostic.Password.SetValue("")
					m.Diagnostic.Password.Blur()
					// Refresh connections with admin privileges
					return m, network.GetConnections()
				}
			}
			return m, nil
		case "esc":
			// Cancel authentication
			m.Network.AuthState = model.AuthNotRequired
			m.Network.AuthMessage = ""
			m.Diagnostic.Password.SetValue("")
			m.Diagnostic.Password.Blur()
			return m, nil
		default:
			// Update password input
			var cmd tea.Cmd
			m.Diagnostic.Password, cmd = m.Diagnostic.Password.Update(msg)
			return m, cmd
		}
	}

	// Handle connectivity mode inputs first
	if m.Network.ConnectivityMode != model.ConnectivityModeNone {
		return m.handleConnectivityInput(msg)
	}

	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "b", "esc":
		m.goBack()
	case "up", "k":
		if m.Network.SelectedItem == model.NetworkTabConnectivity {
			totalContentLines := calculateConnectivityContentLines(m)
			if totalContentLines > m.Network.ConnectivityPerPage {
				if m.Network.ConnectivityPage > 0 {
					m.Network.ConnectivityPage--
				}
				return m, nil
			}
		}
		// Handle table navigation if not used for pagination
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.MoveUp(1)
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.MoveUp(1)
			} else {
				m.Network.DNSTable.MoveUp(1)
			}
		default:
			m.Network.NetworkTable.MoveUp(1)
		}
		return m, nil
	case "shift+tab", "left", "h":
		// Handle tab navigation
		newTab := int(m.Network.SelectedItem) - 1
		if newTab < 0 {
			newTab = len(m.Network.Nav) - 1
		}
		m.Network.SelectedItem = model.ContainerTab(newTab)
		// Set appropriate focus for the new tab
		switch m.Network.SelectedItem {
		case model.NetworkTabInterface:
			m.Network.NetworkTable.Focus()
		case model.NetworkTabConnectivity:
			// Clear focus for connectivity tools
			m.Network.NetworkTable.Blur()
			m.Network.ConnectionsTable.Blur()
			m.Network.RoutesTable.Blur()
			m.Network.DNSTable.Blur()
		case model.NetworkTabConfiguration:
			m.Network.RoutesTable.Focus()
			m.Network.DNSTable.Blur()
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.Focus()
		}
		if m.Network.SelectedItem == model.NetworkTabProtocol && len(m.Network.Connections) == 0 {
			return m, network.GetConnections()
		}
		return m, nil
	case "down", "j":
		// Next page for connectivity results
		if m.Network.SelectedItem == model.NetworkTabConnectivity {
			totalContentLines := calculateConnectivityContentLines(m)
			if totalContentLines > m.Network.ConnectivityPerPage {
				totalPages := (totalContentLines + m.Network.ConnectivityPerPage - 1) / m.Network.ConnectivityPerPage
				if m.Network.ConnectivityPage < totalPages-1 {
					m.Network.ConnectivityPage++
				}
				return m, nil
			}
		}
		// Handle table navigation if not used for pagination
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.MoveDown(1)
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.MoveDown(1)
			} else {
				m.Network.DNSTable.MoveDown(1)
			}
		default:
			m.Network.NetworkTable.MoveDown(1)
		}
		return m, nil
	case "tab", "right", "l":
		// Handle tab navigation
		newTab := int(m.Network.SelectedItem) + 1
		if newTab >= len(m.Network.Nav) {
			newTab = 0
		}
		m.Network.SelectedItem = model.ContainerTab(newTab)
		// Set appropriate focus for the new tab
		switch m.Network.SelectedItem {
		case model.NetworkTabInterface:
			m.Network.NetworkTable.Focus()
		case model.NetworkTabConnectivity:
			// Clear focus for connectivity tools
			m.Network.NetworkTable.Blur()
			m.Network.ConnectionsTable.Blur()
			m.Network.RoutesTable.Blur()
			m.Network.DNSTable.Blur()
		case model.NetworkTabConfiguration:
			m.Network.RoutesTable.Focus()
			m.Network.DNSTable.Blur()
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.Focus()
		}
		if m.Network.SelectedItem == model.NetworkTabProtocol && len(m.Network.Connections) == 0 {
			return m, network.GetConnections()
		}
		return m, nil

	case "1":
		m.Network.SelectedItem = model.NetworkTabInterface
		// Set focus to network interfaces table
		m.Network.NetworkTable.Focus()
		return m, nil
	case "2":
		m.Network.SelectedItem = model.NetworkTabConnectivity
		// Clear any existing focus from tables
		m.Network.NetworkTable.Blur()
		m.Network.ConnectionsTable.Blur()
		m.Network.RoutesTable.Blur()
		m.Network.DNSTable.Blur()
		return m, nil
	case "p":
		if m.Network.SelectedItem == model.NetworkTabConnectivity {
			m.Network.ConnectivityMode = model.ConnectivityModePing
			m.Network.PingInput.Focus()
			return m, nil
		}
	case "t":
		if m.Network.SelectedItem == model.NetworkTabConnectivity {
			m.Network.ConnectivityMode = model.ConnectivityModeTraceroute
			m.Network.TracerouteInput.Focus()
			return m, nil
		}
	case "c":
		if m.Network.SelectedItem == model.NetworkTabConnectivity {
			m.Network.PingResults = nil
			m.Network.TracerouteResults = []network.TracerouteResult{}
			m.Network.ConnectivityPage = 0
			return m, nil
		}
	case "3":
		m.Network.SelectedItem = model.NetworkTabConfiguration
		// Set focus to routes table by default
		m.Network.RoutesTable.Focus()
		m.Network.DNSTable.Blur()
		if len(m.Network.Routes) == 0 && len(m.Network.DNS) == 0 {
			return m, tea.Batch(network.GetRoutes(), network.GetDNS())
		}
		return m, nil
	case "4":
		m.Network.SelectedItem = model.NetworkTabProtocol
		// Set focus to connections table
		m.Network.ConnectionsTable.Focus()
		return m, network.GetConnections()
	case "a":
		// Request authentication for detailed network information
		if m.Network.SelectedItem == model.NetworkTabProtocol && !m.AsRoot && !m.CanRunSudo {
			m.Network.AuthState = model.AuthRequired
			m.Diagnostic.Password.Focus()
			m.Diagnostic.Password.SetValue("")
		}
		return m, nil
	case "pageup":
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.MoveUp(10)
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.MoveUp(10)
			} else {
				m.Network.DNSTable.MoveUp(10)
			}
		default:
			m.Network.NetworkTable.MoveUp(10)
		}
		return m, nil
	case "pagedown":
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.MoveDown(10)
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.MoveDown(10)
			} else {
				m.Network.DNSTable.MoveDown(10)
			}
		default:
			m.Network.NetworkTable.MoveDown(10)
		}
		return m, nil
	case "home":
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.GotoTop()
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.GotoTop()
			} else {
				m.Network.DNSTable.GotoTop()
			}
		default:
			m.Network.NetworkTable.GotoTop()
		}
		return m, nil
	case "end":
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.GotoBottom()
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.GotoBottom()
			} else {
				m.Network.DNSTable.GotoBottom()
			}
		default:
			m.Network.NetworkTable.GotoBottom()
		}
		return m, nil
	case " ":
		if m.Network.SelectedItem == model.NetworkTabConfiguration {
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.Blur()
				m.Network.DNSTable.Focus()
			} else {
				m.Network.DNSTable.Blur()
				m.Network.RoutesTable.Focus()
			}
		}
		return m, nil
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleConnectivityInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		switch m.Network.ConnectivityMode {
		case model.ConnectivityModePing:
			target := m.Network.PingInput.Value()
			if target != "" {
				m.Network.PingInput.SetValue("")
				m.Network.ConnectivityMode = model.ConnectivityModeNone
				m.Network.PingLoading = true
				m.resetSpinner()
				spinnerCmd := m.Ui.Spinner.Tick
				// Force UI refresh by returning a command that will trigger update
				return m, tea.Batch(
					network.Ping(target, 3, m.IsRoot()),
					spinnerCmd,
					func() tea.Msg { return model.ForceRefreshMsg{} },
				)
			}
		case model.ConnectivityModeTraceroute:
			target := m.Network.TracerouteInput.Value()
			if target != "" {
				m.Network.TracerouteInput.SetValue("")
				m.Network.ConnectivityMode = model.ConnectivityModeNone
				m.Network.TracerouteLoading = true
				m.resetSpinner()
				spinnerCmd := m.Ui.Spinner.Tick
				// Force UI refresh by returning a command that will trigger update
				return m, tea.Batch(
					network.Traceroute(target),
					spinnerCmd,
					func() tea.Msg { return model.ForceRefreshMsg{} },
				)
			}
		}
	case "esc":
		m.Network.ConnectivityMode = model.ConnectivityModeNone
		m.Network.PingInput.SetValue("")
		m.Network.TracerouteInput.SetValue("")
		m.Network.PingLoading = false
		m.Network.TracerouteLoading = false
		return m, nil
	default:
		switch m.Network.ConnectivityMode {
		case model.ConnectivityModePing:
			m.Network.PingInput, _ = m.Network.PingInput.Update(msg)
		case model.ConnectivityModeTraceroute:
			m.Network.TracerouteInput, _ = m.Network.TracerouteInput.Update(msg)
		}
	}
	return m, nil
}

func (m Model) handleDiagnosticsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle CPU sub-tab navigation
	if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
		m.Diagnostic.Performance.SelectedItem == model.CPU &&
		m.Diagnostic.Performance.CPUSubTabActive {
		switch msg.String() {
		case "right", "l":
			newTab := int(m.Diagnostic.Performance.CPUSelectedTab) + 1
			if newTab >= 3 { // 3 CPU sub-tabs
				newTab = 0
			}
			m.Diagnostic.Performance.CPUSelectedTab = model.CPUTab(newTab)
			return m, nil
		case "left", "h":
			newTab := int(m.Diagnostic.Performance.CPUSelectedTab) - 1
			if newTab < 0 {
				newTab = 2 // 3 CPU sub-tabs
			}
			m.Diagnostic.Performance.CPUSelectedTab = model.CPUTab(newTab)
			return m, nil
		case "esc", "b":
			// Go back to Performance tab navigation (one level up)
			m.Diagnostic.Performance.CPUSubTabActive = false
			m.Diagnostic.Performance.SubTabNavigationActive = true
			return m, nil
		}
	}

	// Handle Memory sub-tab navigation
	if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
		m.Diagnostic.Performance.SelectedItem == model.Memory &&
		m.Diagnostic.Performance.MemorySubTabActive {
		switch msg.String() {
		case "right", "l":
			newTab := int(m.Diagnostic.Performance.MemorySelectedTab) + 1
			if newTab >= 4 { // 4 Memory sub-tabs
				newTab = 0
			}
			m.Diagnostic.Performance.MemorySelectedTab = model.MemoryTab(newTab)
			return m, nil
		case "left", "h":
			newTab := int(m.Diagnostic.Performance.MemorySelectedTab) - 1
			if newTab < 0 {
				newTab = 3 // 4 Memory sub-tabs
			}
			m.Diagnostic.Performance.MemorySelectedTab = model.MemoryTab(newTab)
			return m, nil
		case "esc", "b":
			// Go back to Performance tab navigation (one level up)
			m.Diagnostic.Performance.MemorySubTabActive = false
			m.Diagnostic.Performance.SubTabNavigationActive = true
			return m, nil
		}
	}

	// Handle Performance main tab navigation
	if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances && m.Diagnostic.Performance.SubTabNavigationActive {
		switch msg.String() {
		case "right", "l":
			newTab := int(m.Diagnostic.Performance.SelectedItem) + 1
			if newTab >= len(m.Diagnostic.Performance.Nav) {
				newTab = 0
			}
			m.Diagnostic.Performance.SelectedItem = model.PerformanceTab(newTab)

			// Auto-load data when switching to I/O, CPU, or Memory tab
			if m.Diagnostic.Performance.SelectedItem == model.InputOutput && m.Diagnostic.Performance.IOMetrics == nil {
				m.Diagnostic.Performance.IOLoading = true
				return m, performance.GetIOMetrics()
			}
			if m.Diagnostic.Performance.SelectedItem == model.CPU && m.Diagnostic.Performance.CPUMetrics == nil {
				m.Diagnostic.Performance.CPULoading = true
				return m, performance.GetCPUMetrics()
			}
			if m.Diagnostic.Performance.SelectedItem == model.Memory && m.Diagnostic.Performance.MemoryMetrics == nil {
				m.Diagnostic.Performance.MemoryLoading = true
				return m, performance.GetMemoryMetrics()
			}
			return m, nil
		case "left", "h":
			newTab := int(m.Diagnostic.Performance.SelectedItem) - 1
			if newTab < 0 {
				newTab = len(m.Diagnostic.Performance.Nav) - 1
			}
			m.Diagnostic.Performance.SelectedItem = model.PerformanceTab(newTab)

			// Auto-load data when switching to I/O, CPU, or Memory tab
			if m.Diagnostic.Performance.SelectedItem == model.InputOutput && m.Diagnostic.Performance.IOMetrics == nil {
				m.Diagnostic.Performance.IOLoading = true
				return m, performance.GetIOMetrics()
			}
			if m.Diagnostic.Performance.SelectedItem == model.CPU && m.Diagnostic.Performance.CPUMetrics == nil {
				m.Diagnostic.Performance.CPULoading = true
				return m, performance.GetCPUMetrics()
			}
			if m.Diagnostic.Performance.SelectedItem == model.Memory && m.Diagnostic.Performance.MemoryMetrics == nil {
				m.Diagnostic.Performance.MemoryLoading = true
				return m, performance.GetMemoryMetrics()
			}
			return m, nil
		case "esc", "b":
			m.Diagnostic.Performance.SubTabNavigationActive = false
			return m, nil
		}
	}
	if m.Diagnostic.LogDetailsView {
		switch msg.String() {
		case "b", "esc", "enter":
			m.Diagnostic.LogDetailsView = false
			m.Diagnostic.SelectedLogEntry = nil
			return m, nil
		case "q", "ctrl+c":
			m.Monitor.ShouldQuit = true
			return m, tea.Quit
		}
		return m, nil
	}
	domain := m.Diagnostic.DomainInput.Value()

	// Handle authentication mode
	if m.Diagnostic.AuthState == model.AuthRequired {
		switch msg.String() {
		case "enter":
			if m.Diagnostic.Password.Value() != "" {
				m.Diagnostic.AuthState = model.AuthInProgress
				m.Diagnostic.AuthMessage = "Authenticating..."
				// Try to authenticate
				err := m.setRoot()
				if err != nil {
					m.Diagnostic.AuthState = model.AuthFailed
					m.Diagnostic.AuthMessage = fmt.Sprintf("Authentication failed: %v", err)
					if !m.SudoAvailable {
						m.Diagnostic.AuthMessage += "\nSudo is not available. Please run as root."
					}
					m.Diagnostic.Password.Reset()
				} else {
					m.Diagnostic.AuthState = model.AuthSuccess
					m.Diagnostic.AuthMessage = "Authentication successful!"
					m.AsRoot = true
					m.Diagnostic.AuthTimer = 2

					// Update SecurityManager with authentication context
					m.Diagnostic.SecurityManager.IsRoot = false // User authenticated via sudo, not actual root
					m.Diagnostic.SecurityManager.CanUseSudo = true
					m.Diagnostic.SecurityManager.SudoPassword = m.Diagnostic.Password.Value()

					m.Diagnostic.Password.SetValue("")
					m.Diagnostic.Password.Blur()
					domain := m.Diagnostic.DomainInput.Value()
					return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
				}
			}
			return m, nil
		case "esc":
			// Cancel authentication
			m.Diagnostic.AuthState = model.AuthNotRequired
			m.Diagnostic.AuthMessage = ""
			m.Diagnostic.Password.SetValue("")
			m.Diagnostic.Password.Blur()
			return m, nil
		default:
			// Update password input
			var cmd tea.Cmd
			m.Diagnostic.Password, cmd = m.Diagnostic.Password.Update(msg)
			return m, cmd
		}
	}

	// Handle domain input mode
	if m.Diagnostic.DomainInputMode {
		switch msg.String() {
		case "enter":
			if domain != "" {
				m.Diagnostic.DomainInputMode = false
				m.Diagnostic.DomainInput.Blur()
				return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
			}
		case "esc":
			// Cancel domain input
			m.Diagnostic.DomainInputMode = false
			m.Diagnostic.DomainInput.Blur()
			return m, nil
		default:
			// Update text input
			var cmd tea.Cmd
			m.Diagnostic.DomainInput, cmd = m.Diagnostic.DomainInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Handle custom time input mode
	if m.Diagnostic.CustomTimeInputMode {
		switch msg.String() {
		case "enter":
			customTime := m.Diagnostic.LogTimeRangeInput.Value()
			if customTime != "" {
				// Validate the custom time input
				if err := m.validateCustomTimeInput(customTime); err != nil {
					m.Diagnostic.CustomTimeInputError = err.Error()
					return m, nil
				}
				// Valid input - apply and load logs
				m.Diagnostic.CustomTimeInputMode = false
				m.Diagnostic.CustomTimeInputError = ""
				m.Diagnostic.LogTimeRangeInput.Blur()
				m.Diagnostic.LogFilters.TimeRange = customTime
				return m, m.loadLogs()
			}
		case "esc":
			// Cancel custom time input
			m.Diagnostic.CustomTimeInputMode = false
			m.Diagnostic.CustomTimeInputError = ""
			m.Diagnostic.LogTimeRangeInput.Blur()
			m.Diagnostic.LogTimeRangeInput.SetValue("")
			return m, nil
		default:
			// Clear error when typing
			m.Diagnostic.CustomTimeInputError = ""
			// Update text input
			var cmd tea.Cmd
			m.Diagnostic.LogTimeRangeInput, cmd = m.Diagnostic.LogTimeRangeInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Handle logs filter input modes
	if m.Diagnostic.LogSearchInput.Focused() {
		switch msg.String() {
		case "esc":
			m.Diagnostic.LogSearchInput.Blur()
			return m, nil
		case "enter":
			m.Diagnostic.LogSearchInput.Blur()
			m.Diagnostic.LogFilters.SearchText = m.Diagnostic.LogSearchInput.Value()
			return m, m.loadLogs()
		default:
			var cmd tea.Cmd
			m.Diagnostic.LogSearchInput, cmd = m.Diagnostic.LogSearchInput.Update(msg)
			return m, cmd
		}
	}

	if m.Diagnostic.LogServiceInput.Focused() {
		switch msg.String() {
		case "esc":
			m.Diagnostic.LogServiceInput.Blur()
			return m, nil
		case "enter":
			m.Diagnostic.LogServiceInput.Blur()
			m.Diagnostic.LogFilters.Service = m.Diagnostic.LogServiceInput.Value()
			return m, m.loadLogs()
		default:
			var cmd tea.Cmd
			m.Diagnostic.LogServiceInput, cmd = m.Diagnostic.LogServiceInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "b", "esc":
		m.goBack()
	case "tab":
		newTab := int(m.Diagnostic.SelectedItem) + 1
		if newTab >= len(m.Diagnostic.Nav) {
			newTab = 0
		}
		m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && m.Diagnostic.LogsInfo == nil {
			return m, m.loadLogs()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances && m.Diagnostic.Performance.HealthMetrics == nil {
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		}
		// Auto-load I/O or CPU data if respective tab is selected and no data exists
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.InputOutput &&
			m.Diagnostic.Performance.IOMetrics == nil {
			m.Diagnostic.Performance.IOLoading = true
			return m, performance.GetIOMetrics()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.CPU &&
			m.Diagnostic.Performance.CPUMetrics == nil {
			m.Diagnostic.Performance.CPULoading = true
			return m, performance.GetCPUMetrics()
		}
		return m, nil
	case "shift+tab":
		// Shift+Tab always switches diagnostic tabs backwards
		newTab := int(m.Diagnostic.SelectedItem) - 1
		if newTab < 0 {
			newTab = len(m.Diagnostic.Nav) - 1
		}
		m.Diagnostic.SelectedItem = model.ContainerTab(newTab)

		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}

		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && m.Diagnostic.LogsInfo == nil {
			return m, m.loadLogs()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances && m.Diagnostic.Performance.HealthMetrics == nil {
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		}
		// Auto-load I/O or CPU data if respective tab is selected and no data exists
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.InputOutput &&
			m.Diagnostic.Performance.IOMetrics == nil {
			m.Diagnostic.Performance.IOLoading = true
			return m, performance.GetIOMetrics()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.CPU &&
			m.Diagnostic.Performance.CPUMetrics == nil {
			m.Diagnostic.Performance.CPULoading = true
			return m, performance.GetCPUMetrics()
		}
		return m, nil
	case "right", "l":
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			// Use left/right for tab navigation like other tabs
			newTab := int(m.Diagnostic.SelectedItem) + 1
			if newTab >= len(m.Diagnostic.Nav) {
				newTab = 0
			}
			m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
			if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
				return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
			}
			if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && m.Diagnostic.LogsInfo == nil {
				return m, m.loadLogs()
			}
			return m, nil
		}

		newTab := int(m.Diagnostic.SelectedItem) + 1
		if newTab >= len(m.Diagnostic.Nav) {
			newTab = 0
		}
		m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && m.Diagnostic.LogsInfo == nil {
			return m, m.loadLogs()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances && m.Diagnostic.Performance.HealthMetrics == nil {
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		}
		// Auto-load I/O or CPU data if respective tab is selected and no data exists
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.InputOutput &&
			m.Diagnostic.Performance.IOMetrics == nil {
			m.Diagnostic.Performance.IOLoading = true
			return m, performance.GetIOMetrics()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.CPU &&
			m.Diagnostic.Performance.CPUMetrics == nil {
			m.Diagnostic.Performance.CPULoading = true
			return m, performance.GetCPUMetrics()
		}
		return m, nil
	case "left", "h":
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {

			newTab := int(m.Diagnostic.SelectedItem) - 1
			if newTab < 0 {
				newTab = len(m.Diagnostic.Nav) - 1
			}
			m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
			if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
				return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
			}
			if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && m.Diagnostic.LogsInfo == nil {
				return m, m.loadLogs()
			}
			return m, nil
		}

		newTab := int(m.Diagnostic.SelectedItem) - 1
		if newTab < 0 {
			newTab = len(m.Diagnostic.Nav) - 1
		}
		m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && m.Diagnostic.LogsInfo == nil {
			return m, m.loadLogs()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances && m.Diagnostic.Performance.HealthMetrics == nil {
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		}
		// Auto-load I/O or CPU data if respective tab is selected and no data exists
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.InputOutput &&
			m.Diagnostic.Performance.IOMetrics == nil {
			m.Diagnostic.Performance.IOLoading = true
			return m, performance.GetIOMetrics()
		}
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances &&
			m.Diagnostic.Performance.SelectedItem == model.CPU &&
			m.Diagnostic.Performance.CPUMetrics == nil {
			m.Diagnostic.Performance.CPULoading = true
			return m, performance.GetCPUMetrics()
		}
		return m, nil
	case "shift+right", "shift+l":
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			if m.Diagnostic.LogFilterSelected == 0 {
				timeRanges := []string{"All", "5m", "1h", "24h", "7d", "Custom"}
				if m.Diagnostic.LogTimeSelected < len(timeRanges)-1 {
					m.Diagnostic.LogTimeSelected++
					m.applyTimeRangeSelection()
				}
			} else {
				levels := []string{"All", "Error", "Warn", "Info", "Debug"}
				if m.Diagnostic.LogLevelSelected < len(levels)-1 {
					m.Diagnostic.LogLevelSelected++
					m.applyLogLevelSelection()
				}
			}
			return m, nil
		}
		return m, nil
	case "shift+left", "shift+h":
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			if m.Diagnostic.LogFilterSelected == 0 {
				if m.Diagnostic.LogTimeSelected > 0 {
					m.Diagnostic.LogTimeSelected--
					m.applyTimeRangeSelection()
				}
			} else {
				if m.Diagnostic.LogLevelSelected > 0 {
					m.Diagnostic.LogLevelSelected--
					m.applyLogLevelSelection()
				}
			}
			return m, nil
		}
		return m, nil
	case "1":
		m.Diagnostic.SelectedItem = model.DiagnosticSecurityChecks
		if len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		return m, nil
	case "2":
		m.Diagnostic.SelectedItem = model.DiagnosticTabPerformances
		return m, nil
	case "3":
		m.Diagnostic.SelectedItem = model.DiagnosticTabLogs
		if m.Diagnostic.LogsInfo == nil {
			return m, m.loadLogs()
		}
		return m, nil
	case "up", "k":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			m.Diagnostic.SecurityTable.MoveUp(1)
		case model.DiagnosticTabLogs:
			m.Diagnostic.LogsTable.MoveUp(1)
		default:
			m.Diagnostic.DiagnosticTable.MoveUp(1)
		}
		return m, nil
	case "down", "j":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			m.Diagnostic.SecurityTable.MoveDown(1)
		case model.DiagnosticTabLogs:
			m.Diagnostic.LogsTable.MoveDown(1)
		default:
			m.Diagnostic.DiagnosticTable.MoveDown(1)
		}
		return m, nil
	case "pageup":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			m.Diagnostic.SecurityTable.MoveUp(10)
		case model.DiagnosticTabLogs:
			m.Diagnostic.LogsTable.MoveUp(10)
		default:
			m.Diagnostic.DiagnosticTable.MoveUp(10)
		}
		return m, nil
	case "pagedown":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			m.Diagnostic.SecurityTable.MoveDown(10)
		case model.DiagnosticTabLogs:
			m.Diagnostic.LogsTable.MoveDown(10)
		default:
			m.Diagnostic.DiagnosticTable.MoveDown(10)
		}
		return m, nil
	case "home":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			m.Diagnostic.SecurityTable.GotoTop()
		case model.DiagnosticTabLogs:
			m.Diagnostic.LogsTable.GotoTop()
		default:
			m.Diagnostic.DiagnosticTable.GotoTop()
		}
		return m, nil
	case "end":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			m.Diagnostic.SecurityTable.GotoBottom()
		case model.DiagnosticTabLogs:
			m.Diagnostic.LogsTable.GotoBottom()
		default:
			m.Diagnostic.DiagnosticTable.GotoBottom()
		}
		return m, nil
	case "r":
		switch m.Diagnostic.SelectedItem {
		case model.DiagnosticSecurityChecks:
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		case model.DiagnosticTabLogs:
			return m, m.loadLogs()
		case model.DiagnosticTabPerformances:
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		}
	case "a":
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && !m.AsRoot && !m.CanRunSudo {
			m.Diagnostic.AuthState = model.AuthRequired
			m.Diagnostic.AuthMessage = "Enter password for admin access:"
			m.Diagnostic.Password.Focus()
			m.Diagnostic.Password.SetValue("")
		}
		return m, nil
	case "d":
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.DomainInputMode = true
			m.Diagnostic.DomainInput.Focus()
			return m, nil
		}
		// Show log entry details when on logs tab
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			if m.Diagnostic.LogsInfo != nil && len(m.Diagnostic.LogsInfo.Entries) > 0 {
				selectedRowIndex := m.Diagnostic.LogsTable.Cursor()
				if selectedRowIndex >= 0 && selectedRowIndex < len(m.Diagnostic.LogsInfo.Entries) {
					m.Diagnostic.SelectedLogEntry = &m.Diagnostic.LogsInfo.Entries[selectedRowIndex]
					m.Diagnostic.LogDetailsView = true
				}
			}
			return m, nil
		}
		return m, nil
	case "/":
		// Focus search input when on logs tab
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			m.Diagnostic.LogSearchInput.Focus()
			return m, nil
		}
	case "s":
		// Focus service filter when on logs tab
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs && !m.Diagnostic.LogSearchInput.Focused() {
			m.Diagnostic.LogServiceInput.Focus()
			return m, nil
		}
	case " ":
		// On logs tab, space switches between Time and Level filter
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			m.Diagnostic.LogFilterSelected = (m.Diagnostic.LogFilterSelected + 1) % 2
			return m, nil
		}
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	case "enter":
		if m.Diagnostic.SelectedItem == model.DiagnosticTabPerformances {
			// If already in Performance sub-tab navigation and on CPU tab, activate CPU sub-tab navigation
			if m.Diagnostic.Performance.SubTabNavigationActive && m.Diagnostic.Performance.SelectedItem == model.CPU {
				m.Diagnostic.Performance.SubTabNavigationActive = false
				m.Diagnostic.Performance.CPUSubTabActive = true
				return m, nil
			}
			// If already in Performance sub-tab navigation and on Memory tab, activate Memory sub-tab navigation
			if m.Diagnostic.Performance.SubTabNavigationActive && m.Diagnostic.Performance.SelectedItem == model.Memory {
				m.Diagnostic.Performance.SubTabNavigationActive = false
				m.Diagnostic.Performance.MemorySubTabActive = true
				return m, nil
			}
			// Otherwise activate Performance sub-tab navigation
			m.Diagnostic.Performance.SubTabNavigationActive = true
			return m, nil
		}
		// On logs tab, check if Custom time range is selected
		if m.Diagnostic.SelectedItem == model.DiagnosticTabLogs {
			// If "Custom" is selected, enter custom time input mode
			if m.Diagnostic.LogTimeSelected == 5 { // Custom is at index 5
				m.Diagnostic.CustomTimeInputMode = true
				m.Diagnostic.LogTimeRangeInput.Focus()
				m.Diagnostic.LogTimeRangeInput.SetValue("")
				return m, nil
			}
			// Otherwise, reload logs with current filters
			return m, m.loadLogs()
		}
		// Check if we're on security checks tab and get selected security check
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityTable.SelectedRow()) > 0 {
			selectedRow := m.Diagnostic.SecurityTable.SelectedRow()
			if len(selectedRow) > 0 {
				checkName := selectedRow[0]

				// Check if admin privileges are required for this diagnostic
				if !m.canAccessDiagnostic(checkName) {
					m.Diagnostic.AuthState = model.AuthRequired
					m.Diagnostic.AuthMessage = fmt.Sprintf("Admin privileges required for: %s\nEnter password:", checkName)
					m.Diagnostic.Password.Focus()
					m.Diagnostic.Password.SetValue("")
					return m, nil
				}

				// Execute the diagnostic check
				switch checkName {
				case "SSL Certificate":
					return m, m.Diagnostic.SecurityManager.RunCertificateDisplay()
				case "SSH Root Login":
					return m, m.Diagnostic.SecurityManager.DisplaySSHRootInfos()
				case "Open Ports":
					return m, m.Diagnostic.SecurityManager.DisplayOpenedPortsInfos()
				case "Firewall Status":
					return m, m.Diagnostic.SecurityManager.DisplayFirewallInfos()
				case "Auto Ban":
					return m, m.Diagnostic.SecurityManager.DisplayAutoBanInfos()
				}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleSecurityCheckMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case security.SecurityMsg:
		checks := []security.SecurityCheck(msg)
		m.Diagnostic.SecurityChecks = checks

		// Update authentication state based on root/sudo availability
		if !m.AsRoot && !m.CanRunSudo {
			// Check if any admin-required checks are present
			adminChecks := m.getAdminRequiredChecks()
			hasAdminChecks := false
			for _, check := range checks {
				if slices.Contains(adminChecks, check.Name) {
					hasAdminChecks = true
				}
				if hasAdminChecks {
					break
				}
			}

			if hasAdminChecks && m.Diagnostic.AuthState == model.AuthNotRequired {
				m.Diagnostic.AuthMessage = "Some checks require admin privileges. Press 'a' to authenticate."
			}
		}

		return m, m.updateSecurityTable()
	}
	return m, nil
}

// ------------------------- handler for certificate display messages -------------------------
func (m Model) handleCertificateDisplayMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case security.CertificateDisplayMsg:
		certInfo := security.CertificateInfos(msg)
		// Store certificate info in model for display
		m.Diagnostic.CertificateInfo = &certInfo
		// Switch to certificate details view
		m.setState(model.StateCertificateDetails)
		return m, nil
	}
	return m, nil
}

func (m Model) handleCertificateDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "up", "k":
		m.Ui.Viewport.ScrollUp(1)
	case "down", "j":
		m.Ui.Viewport.ScrollDown(1)
	case "pageup":
		m.Ui.Viewport.PageUp()
	case "pagedown":
		m.Ui.Viewport.PageDown()
	case "home":
		m.Ui.Viewport.GotoTop()
	case "end":
		m.Ui.Viewport.GotoBottom()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

// ------------------------- handle SSH root display messages -------------------------
func (m Model) handleSSHRootDisplayMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case security.SSHRootMsg:
		sshRootInfo := security.SSHRootInfos(msg)
		// Store SSH root info in model for display
		m.Diagnostic.SSHRootInfo = &sshRootInfo
		// Switch to SSH root details view
		m.setState(model.StateSSHRootDetails)
		return m, nil
	}
	return m, nil
}

func (m Model) handleSSHRootDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "up", "k":
		m.Ui.Viewport.ScrollUp(1)
	case "down", "j":
		m.Ui.Viewport.ScrollDown(1)
	case "pageup":
		m.Ui.Viewport.PageUp()
	case "pagedown":
		m.Ui.Viewport.PageDown()
	case "home":
		m.Ui.Viewport.GotoTop()
	case "end":
		m.Ui.Viewport.GotoBottom()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

// ------------------------- handler for opened ports display messages -------------------------
func (m Model) handleOpenedPortsDisplayMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case security.OpenedPortsMsg:
		openedPortsInfo := security.OpenedPortsInfos(msg)
		// Store opened ports info in model for display
		m.Diagnostic.OpenedPortsInfo = &openedPortsInfo
		// Switch to opened ports details view
		m.setState(model.StateOpenedPortsDetails)
		return m, m.updatePortsTable()
	}
	return m, nil
}

// ------------------------- handler for firewall display messages -------------------------
func (m Model) handleFirewallDisplayMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case security.FirewallMsg:
		firewallInfo := security.FirewallInfos(msg)
		// Store firewall info in model for display
		m.Diagnostic.FirewallInfo = &firewallInfo
		// Switch to firewall details view
		m.setState(model.StateFirewallDetails)
		return m, m.updateFirewallTable()
	}
	return m, nil
}

// ------------------------- handler for autoban display messages -------------------------
func (m Model) handleAutoBanDisplayMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case security.AutoBanMsg:
		autoBanInfo := security.AutoBanInfos(msg)
		// Store auto-ban info in model for display
		m.Diagnostic.AutoBanInfo = &autoBanInfo
		// Switch to auto-ban details view
		m.setState(model.StateAutoBanDetails)
		return m, m.updateAutoBanTable()
	}
	return m, nil
}

func (m Model) handleOpenedPortsDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "up", "k":
		m.Diagnostic.PortsTable.MoveUp(1)
	case "down", "j":
		m.Diagnostic.PortsTable.MoveDown(1)
	case "pageup":
		m.Diagnostic.PortsTable.MoveUp(10)
	case "pagedown":
		m.Diagnostic.PortsTable.MoveDown(10)
	case "home":
		m.Diagnostic.PortsTable.GotoTop()
	case "end":
		m.Diagnostic.PortsTable.GotoBottom()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleFirewallDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "up", "k":
		m.Diagnostic.FirewallTable.MoveUp(1)
	case "down", "j":
		m.Diagnostic.FirewallTable.MoveDown(1)
	case "pageup":
		m.Diagnostic.FirewallTable.MoveUp(10)
	case "pagedown":
		m.Diagnostic.FirewallTable.MoveDown(10)
	case "home":
		m.Diagnostic.FirewallTable.GotoTop()
	case "end":
		m.Diagnostic.FirewallTable.GotoBottom()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleAutoBanDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "up", "k":
		m.Diagnostic.AutoBanTable.MoveUp(1)
	case "down", "j":
		m.Diagnostic.AutoBanTable.MoveDown(1)
	case "pageup":
		m.Diagnostic.AutoBanTable.MoveUp(10)
	case "pagedown":
		m.Diagnostic.AutoBanTable.MoveDown(10)
	case "home":
		m.Diagnostic.AutoBanTable.GotoTop()
	case "end":
		m.Diagnostic.AutoBanTable.GotoBottom()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleReportingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {

	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

// ------------------------- General keys & shortcuts -------------------------

func (m Model) handleContainerMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "up", "k":
		if m.Monitor.SelectedMenuItem > 0 {
			m.Monitor.SelectedMenuItem--
		}
		return m, nil
	case "down", "j":
		if m.Monitor.SelectedMenuItem < len(m.Monitor.ContainerMenuItems)-1 {
			m.Monitor.SelectedMenuItem++
		}
		return m, nil
	case "enter":
		return m.executeContainerMenuAction()
	case "o":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.setState(model.StateContainer)
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "l":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.setState(model.StateContainerLogs)
		m.Monitor.ContainerLogsLoading = true
		if m.Monitor.ContainerLogsStreaming {
			return m, m.Monitor.App.StopLogsStreamCmd(m.Monitor.SelectedContainer.ID)
		}
		return m, m.Monitor.App.GetContainerLogsCmd(
			m.Monitor.SelectedContainer.ID)
	case "r":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "restart"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "d":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "delete"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "x":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "remove"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "s":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerStateCmd(m.Monitor.SelectedContainer.ID)
	case "p":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerPauseCmd(m.Monitor.SelectedContainer.ID)
	case "e":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: m.Monitor.SelectedContainer.ID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case "c":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.LastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	case "esc", "b":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.Monitor.SelectedContainer = nil
		return m, nil
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) executeContainerMenuAction() (tea.Model, tea.Cmd) {
	if m.Monitor.SelectedMenuItem >= len(m.Monitor.ContainerMenuItems) {
		return m, nil
	}
	m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
	action := m.Monitor.ContainerMenuItems[m.Monitor.SelectedMenuItem].Action
	switch action {
	case "open_single":
		m.setState(model.StateContainer)
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "logs":
		m.setState(model.StateContainerLogs)
		m.Monitor.ContainerLogsLoading = true
		return m, m.Monitor.App.GetContainerLogsCmd(
			m.Monitor.SelectedContainer.ID)
	case "restart":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "restart"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "delete":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "delete"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "remove":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "remove"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "toggle_start":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerStateCmd(m.Monitor.SelectedContainer.ID)
	case "toggle_pause":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerPauseCmd(m.Monitor.SelectedContainer.ID)
	case "exec":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: m.Monitor.SelectedContainer.ID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case "commit":
		m.Monitor.ContainerMenuState = model.ContainerMenuState(0) // ContainerMenuHidden
		m.LastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	}
	return m, nil
}

// ------------------------- Confirmation box keys -------------------------

func (m Model) handleConfirmationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.HelpSystem.ToggleHelp()
	case "y", "Y":
		m.ConfirmationVisible = false
		switch m.ConfirmationAction {
		case "delete":
			if containerID, ok := m.ConfirmationData.(string); ok {
				m.OperationInProgress = true
				m.ConfirmationAction = ""
				m.ConfirmationData = nil
				return m, m.Monitor.App.DeleteContainerCmd(containerID, false)
			}
		case "remove":
			if containerID, ok := m.ConfirmationData.(string); ok {
				m.OperationInProgress = true
				m.ConfirmationAction = ""
				m.ConfirmationData = nil
				return m, m.Monitor.App.DeleteContainerCmd(containerID, true)
			}
		case "restart":
			if containerID, ok := m.ConfirmationData.(string); ok {
				m.OperationInProgress = true
				m.ConfirmationAction = ""
				m.ConfirmationData = nil
				return m, m.Monitor.App.RestartContainerCmd(containerID)
			}
		}
		return m, nil
	case "n", "N", "esc":
		m.ConfirmationVisible = false
		m.ConfirmationMessage = ""
		m.ConfirmationAction = ""
		m.ConfirmationData = nil
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

// ------------------------- Search mode -------------------------

func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		var tcmd tea.Cmd
		m.Ui.SearchMode = false
		m.Monitor.ProcessTable.Focus()
		m.Monitor.Container.Focus()
		if m.Ui.SelectedMonitor == 1 {
			tcmd = m.updateProcessTable()
		} else {
			tcmd = m.Monitor.App.UpdateApp()
		}
		return m, tcmd
	default:
		var cmd tea.Cmd
		m.Ui.SearchInput, cmd = m.Ui.SearchInput.Update(msg)
		return m, cmd
	}
}

// ------------------------- Tick -------------------------

func (m Model) handleTickMsg() (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{
		tick(),
		info.UpdateSystemInfo(),
		resource.UpdateCPUInfo(),
		resource.UpdateMemoryInfo(),
		resource.UpdateDiskInfo(),
		resource.UpdateNetworkInfo(),
		proc.UpdateProcesses(),
		m.Monitor.App.UpdateApp(),
		// Refresh network connectivity data when in network view
		func() tea.Cmd {
			if m.Ui.State == model.StateNetwork {
				switch m.Network.SelectedItem {
				case model.NetworkTabProtocol:
					return network.GetConnections()
				case model.NetworkTabConfiguration:
					return tea.Batch(network.GetRoutes(), network.GetDNS())
				}
			}
			return nil
		}(),
	}

	if m.Diagnostic.AuthState == model.AuthSuccess && m.Diagnostic.AuthTimer > 0 {
		m.Diagnostic.AuthTimer--
		if m.Diagnostic.AuthTimer == 0 {
			m.Diagnostic.AuthState = model.AuthNotRequired
			m.Diagnostic.AuthMessage = ""
		}
	}

	if m.Network.AuthState == model.AuthSuccess && m.Network.AuthTimer > 0 {
		m.Network.AuthTimer--
		if m.Network.AuthTimer == 0 {
			m.Network.AuthState = model.AuthNotRequired
			m.Network.AuthMessage = ""
		}
	}

	m.updateCharts()
	updateNetworkCmd := m.updateNetworkTable()
	if updateNetworkCmd != nil {
		cmds = append(cmds, updateNetworkCmd)
	}
	return m, tea.Batch(cmds...)
}

// ------------------------- Progress frame updates -------------------------

func (m Model) handleProgressFrame(msg progress.FrameMsg) (tea.Model, tea.Cmd) {
	var (
		progCmd tea.Cmd
		cmds    []tea.Cmd
	)

	updatedModel, progCmd := m.Monitor.CpuProgress.Update(msg)
	m.Monitor.CpuProgress = updatedModel.(progress.Model)
	cmds = append(cmds, progCmd)

	updatedModel, progCmd = m.Monitor.MemProgress.Update(msg)
	m.Monitor.MemProgress = updatedModel.(progress.Model)
	cmds = append(cmds, progCmd)

	updatedModel, progCmd = m.Monitor.SwapProgress.Update(msg)
	m.Monitor.SwapProgress = updatedModel.(progress.Model)
	cmds = append(cmds, progCmd)

	for key, p := range m.Monitor.DiskProgress {
		updatedModel, progCmd := (p).Update(msg)
		newModel := updatedModel.(progress.Model)
		m.Monitor.DiskProgress[key] = newModel
		cmds = append(cmds, progCmd)
	}
	return m, tea.Batch(cmds...)
}

// ------------------------- Mouse/trackpad handling -------------------------

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if !m.MouseEnabled {
		return m, nil
	}
	mouseEvent := tea.MouseEvent(msg)
	if time.Since(m.LastScrollTime) < 60*time.Millisecond {
		return m, nil
	}
	m.LastScrollTime = time.Now()

	switch {
	case mouseEvent.Button == tea.MouseButtonWheelUp &&
		mouseEvent.Action == tea.MouseActionPress:
		return m.handleScrollUp()
	case mouseEvent.Button == tea.MouseButtonWheelDown &&
		mouseEvent.Action == tea.MouseActionPress:
		return m.handleScrollDown()
	}

	return m, nil
}

func (m Model) handleScrollUp() (tea.Model, tea.Cmd) {
	switch m.Ui.State {
	case model.StateSystem, model.StateContainer,
		model.StateDiagnostics,
		model.StateCertificateDetails,
		model.StateSSHRootDetails,
		model.StateOpenedPortsDetails,
		model.StateFirewallDetails,
		model.StateAutoBanDetails,
		model.StatePerformance,
		model.StateSystemHealth,
		model.StateInputOutput,
		model.StateCPU,
		model.StateMemory,
		model.StateQuickTests,
		model.StateReporting:
		m.Ui.Viewport.ScrollUp(m.ScrollSensitivity)
	case model.StateContainerLogs:
		m.LogsViewport.ScrollUp(m.ScrollSensitivity)
	case model.StateProcess:
		m.Monitor.ProcessTable.MoveUp(m.ScrollSensitivity)
	case model.StateContainers:
		m.Monitor.Container.MoveUp(m.ScrollSensitivity)
	case model.StateNetwork:
		switch m.Network.SelectedItem {
		case model.NetworkTabConnectivity:
			m.Ui.Viewport.ScrollUp(m.ScrollSensitivity)
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.MoveUp(m.ScrollSensitivity)
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.MoveUp(m.ScrollSensitivity)
			} else {
				m.Network.DNSTable.MoveUp(m.ScrollSensitivity)
			}
		default:
			m.Network.NetworkTable.MoveUp(m.ScrollSensitivity)
		}
	}
	return m, nil
}

func (m Model) handleScrollDown() (tea.Model, tea.Cmd) {
	switch m.Ui.State {
	case model.StateSystem, model.StateContainer,
		model.StateDiagnostics,
		model.StateCertificateDetails,
		model.StateSSHRootDetails,
		model.StateOpenedPortsDetails,
		model.StateFirewallDetails,
		model.StateAutoBanDetails,
		model.StatePerformance,
		model.StateSystemHealth,
		model.StateInputOutput,
		model.StateCPU,
		model.StateMemory,
		model.StateQuickTests,
		model.StateReporting:
		m.Ui.Viewport.ScrollDown(m.ScrollSensitivity)
	case model.StateContainerLogs:
		m.LogsViewport.ScrollDown(m.ScrollSensitivity)
	case model.StateProcess:
		m.Monitor.ProcessTable.MoveDown(m.ScrollSensitivity)
	case model.StateContainers:
		m.Monitor.Container.MoveDown(m.ScrollSensitivity)
	case model.StateNetwork:
		switch m.Network.SelectedItem {
		case model.NetworkTabProtocol:
			m.Network.ConnectionsTable.MoveDown(m.ScrollSensitivity)
		case model.NetworkTabConfiguration:
			if m.Network.RoutesTable.Focused() {
				m.Network.RoutesTable.MoveDown(m.ScrollSensitivity)
			} else {
				m.Network.DNSTable.MoveDown(m.ScrollSensitivity)
			}
		default:
			m.Network.NetworkTable.MoveDown(m.ScrollSensitivity)
		}
	}
	return m, nil
}

func (m Model) handleNetworkMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case network.ConnectionsMsg:
		m.Network.Connections = msg
		return m, m.updateConnectionsTable()
	case network.RoutesMsg:
		m.Network.Routes = msg
		return m, m.updateRoutesTable()
	case network.DNSMsg:
		m.Network.DNS = msg
		return m, m.updateDNSTable()
	case network.PingMsg:
		m.Network.PingResults = append(m.Network.PingResults, network.PingResult(msg))
		m.Network.PingLoading = false
		return m, nil
	case network.TracerouteMsg:
		m.Network.TracerouteResults = append(m.Network.TracerouteResults, network.TracerouteResult(msg))
		m.Network.TracerouteLoading = false
		m.Network.ConnectivityPage = 0
		return m, nil
	}
	return m, nil
}

// ------------------------- handler for logs display messages -------------------------
func (m Model) handleLogsDisplayMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logs.LogsMsg:
		logsInfo := logs.LogsInfos(msg)
		// Store logs info in model for display
		m.Diagnostic.LogsInfo = &logsInfo
		return m, m.updateLogsTable()
	}
	return m, nil
}

// validateCustomTimeInput validates user input for custom time ranges
func (m Model) validateCustomTimeInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("time range cannot be empty")
	}

	// Check for common patterns that journalctl accepts
	validPatterns := []string{
		// Relative time with "ago"
		"ago", "minute", "hour", "day", "week", "month", "year",
		// Absolute dates (basic check for digits and dashes/colons)
		"-", ":", // Date separators
	}

	inputLower := strings.ToLower(input)

	// Check if input contains at least one valid pattern or looks like a date
	hasValidPattern := false
	for _, pattern := range validPatterns {
		if strings.Contains(inputLower, pattern) {
			hasValidPattern = true
			break
		}
	}

	// Also accept if it looks like a date (contains digits)
	hasDigits := false
	for _, char := range input {
		if char >= '0' && char <= '9' {
			hasDigits = true
			break
		}
	}

	if !hasValidPattern && !hasDigits {
		return fmt.Errorf("invalid time format. Examples:\n  • '5 minutes ago', '2 hours ago', '3 days ago'\n  • '2025-01-08' or '2025-01-08 14:30:00'\n  • 'today', 'yesterday'")
	}

	// Additional validation: if it contains "ago", it should have a number before it
	if strings.Contains(inputLower, "ago") {
		words := strings.Fields(input)
		if len(words) < 2 {
			return fmt.Errorf("invalid format with 'ago'. Use: '<number> <unit> ago'\n  Examples: '5 minutes ago', '2 hours ago'")
		}
		// Check if first word is a number
		_, err := strconv.Atoi(words[0])
		if err != nil {
			return fmt.Errorf("expected a number before time unit.\n  Examples: '5 minutes ago', '2 hours ago'")
		}
	}

	return nil
}

func (m Model) handlePerformanceKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		switch m.Diagnostic.Performance.SelectedItem {
		case model.SystemHealth:
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		case model.InputOutput:
			m.Diagnostic.Performance.IOLoading = true
			return m, performance.GetIOMetrics()
		case model.CPU:
			m.Diagnostic.Performance.CPULoading = true
			return m, performance.GetCPUMetrics()
		default:
			m.Diagnostic.Performance.HealthLoading = true
			return m, performance.GetHealthMetrics()
		}
	default:
		return m.handleGeneralKeys(msg)
	}
}
