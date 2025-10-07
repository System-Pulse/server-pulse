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
	proc "github.com/System-Pulse/server-pulse/system/process"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/system/security"
	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	if msg.Width < 40 || msg.Height < 10 {
		m.Ui.Ready = false
		return m, nil
	}
	m.Ui.Ready = true

	headerHeight := lipgloss.Height(m.renderHeader())
	navHeight := lipgloss.Height(m.renderCurrentNav())
	footerHeight := lipgloss.Height(m.renderFooter())

	verticalMargin := headerHeight + navHeight + footerHeight
	contentHeight := max(1, msg.Height-verticalMargin)

	m.Ui.Viewport.Width = msg.Width
	m.Ui.Viewport.Height = contentHeight

	m.LogsViewport.Width = msg.Width
	m.LogsViewport.Height = contentHeight

	tableHeight := max(1, contentHeight-3)

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
	if m.Monitor.ContainerMenuState == v.ContainerMenuVisible {
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
	case model.StateReporting:
		return m.handleReportingKeys(msg)
	}

	return m, nil
}

func (m Model) handleHomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
			m.Monitor.ContainerMenuState = v.ContainerMenuVisible
			m.Monitor.SelectedMenuItem = 0
		}
	default:
		return m.handleGeneralKeys(msg)
	}
	return m, nil
}

func (m Model) handleContainerSingleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "tab", "right", "l":
		newTab := int(m.Network.SelectedItem) + 1
		if newTab >= len(m.Network.Nav) {
			newTab = 0
		}
		m.Network.SelectedItem = model.ContainerTab(newTab)
		return m, nil
	case "shift+tab", "left", "h":
		newTab := int(m.Network.SelectedItem) - 1
		if newTab < 0 {
			newTab = len(m.Network.Nav) - 1
		}
		m.Network.SelectedItem = model.ContainerTab(newTab)
		return m, nil
	case "1":
		m.Network.SelectedItem = model.NetworkTabInterface
		return m, nil
	case "2":
		m.Network.SelectedItem = model.NetworkTabConnectivity
		return m, nil
	case "3":
		m.Network.SelectedItem = model.NetworkTabConfiguration
		return m, nil
	case "4":
		m.Network.SelectedItem = model.NetworkTabProtocol
		return m, nil
	case "up", "k":
		m.Network.NetworkTable.MoveUp(1)
		return m, nil
	case "down", "j":
		m.Network.NetworkTable.MoveDown(1)
		return m, nil
	case "pageup":
		m.Network.NetworkTable.MoveUp(10)
		return m, nil
	case "pagedown":
		m.Network.NetworkTable.MoveDown(10)
		return m, nil
	case "home":
		m.Network.NetworkTable.GotoTop()
		return m, nil
	case "end":
		m.Network.NetworkTable.GotoBottom()
		return m, nil
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleDiagnosticsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
					if !m.Diagnostic.SudoAvailable {
						m.Diagnostic.AuthMessage += "\nSudo is not available. Please run as root."
					}
					m.Diagnostic.Password.Reset()
				} else {
					m.Diagnostic.AuthState = model.AuthSuccess
					m.Diagnostic.AuthMessage = "Authentication successful!"
					m.Diagnostic.IsRoot = true
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

	switch msg.String() {
	case "b", "esc":
		m.goBack()
	case "tab", "right", "l":
		newTab := int(m.Diagnostic.SelectedItem) + 1
		if newTab >= len(m.Diagnostic.Nav) {
			newTab = 0
		}
		m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
		// Auto-load security checks if switching to security tab and not loaded
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
			// domain := m.Diagnostic.DomainInput.Value()
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		return m, nil
	case "shift+tab", "left", "h":
		newTab := int(m.Diagnostic.SelectedItem) - 1
		if newTab < 0 {
			newTab = len(m.Diagnostic.Nav) - 1
		}
		m.Diagnostic.SelectedItem = model.ContainerTab(newTab)
		// Auto-load security checks if switching to security tab and not loaded
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		return m, nil
	case "1":
		m.Diagnostic.SelectedItem = model.DiagnosticSecurityChecks
		// Auto-load security checks if not already loaded
		if len(m.Diagnostic.SecurityChecks) == 0 {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
		return m, nil
	case "2":
		m.Diagnostic.SelectedItem = model.DiagnosticTabPerformances
		return m, nil
	case "3":
		m.Diagnostic.SelectedItem = model.DiagnosticTabLogs
		return m, nil
	case "up", "k":
		// Handle navigation based on current diagnostic tab
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.SecurityTable.MoveUp(1)
		} else {
			m.Diagnostic.DiagnosticTable.MoveUp(1)
		}
		return m, nil
	case "down", "j":
		// Handle navigation based on current diagnostic tab
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.SecurityTable.MoveDown(1)
		} else {
			m.Diagnostic.DiagnosticTable.MoveDown(1)
		}
		return m, nil
	case "pageup":
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.SecurityTable.MoveUp(10)
		} else {
			m.Diagnostic.DiagnosticTable.MoveUp(10)
		}
		return m, nil
	case "pagedown":
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.SecurityTable.MoveDown(10)
		} else {
			m.Diagnostic.DiagnosticTable.MoveDown(10)
		}
		return m, nil
	case "home":
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.SecurityTable.GotoTop()
		} else {
			m.Diagnostic.DiagnosticTable.GotoTop()
		}
		return m, nil
	case "end":
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.SecurityTable.GotoBottom()
		} else {
			m.Diagnostic.DiagnosticTable.GotoBottom()
		}
		return m, nil
	case "r":
		// Refresh security checks when on security tab
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			return m, m.Diagnostic.SecurityManager.RunSecurityChecks(domain)
		}
	case "a":
		// Request authentication for admin checks
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks && !m.Diagnostic.IsRoot && !m.Diagnostic.CanRunSudo {
			m.Diagnostic.AuthState = model.AuthRequired
			m.Diagnostic.AuthMessage = "Enter password for admin access:"
			m.Diagnostic.Password.Focus()
			m.Diagnostic.Password.SetValue("")
		}
		return m, nil
	case "d":
		// Enter domain input mode when on security tab
		if m.Diagnostic.SelectedItem == model.DiagnosticSecurityChecks {
			m.Diagnostic.DomainInputMode = true
			m.Diagnostic.DomainInput.Focus()
			return m, nil
		}
		return m, nil
	case "q", "ctrl+c":
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	case "enter":
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
		if !m.Diagnostic.IsRoot && !m.Diagnostic.CanRunSudo {
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
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.setState(model.StateContainer)
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "l":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.setState(model.StateContainerLogs)
		m.Monitor.ContainerLogsLoading = true
		if m.Monitor.ContainerLogsStreaming {
			return m, m.Monitor.App.StopLogsStreamCmd(m.Monitor.SelectedContainer.ID)
		}
		return m, m.Monitor.App.GetContainerLogsCmd(
			m.Monitor.SelectedContainer.ID)
	case "r":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "restart"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "d":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "delete"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "x":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "remove"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "s":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerStateCmd(m.Monitor.SelectedContainer.ID)
	case "p":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerPauseCmd(m.Monitor.SelectedContainer.ID)
	case "e":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: m.Monitor.SelectedContainer.ID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case "c":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.LastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	case "esc", "b":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
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
	m.Monitor.ContainerMenuState = v.ContainerMenuHidden
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
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "restart"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "delete":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "delete"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "remove":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "remove"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "toggle_start":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerStateCmd(m.Monitor.SelectedContainer.ID)
	case "toggle_pause":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerPauseCmd(m.Monitor.SelectedContainer.ID)
	case "exec":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: m.Monitor.SelectedContainer.ID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case "commit":
		m.Monitor.ContainerMenuState = v.ContainerMenuHidden
		m.LastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	}
	return m, nil
}

// ------------------------- Confirmation box keys -------------------------

func (m Model) handleConfirmationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
	}

	// Gestion du timer d'authentification
	if m.Diagnostic.AuthState == model.AuthSuccess && m.Diagnostic.AuthTimer > 0 {
		m.Diagnostic.AuthTimer--
		if m.Diagnostic.AuthTimer == 0 {
			m.Diagnostic.AuthState = model.AuthNotRequired
			m.Diagnostic.AuthMessage = ""
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
	case model.StateSystem, model.StateContainerLogs:
		m.Ui.Viewport.ScrollUp(m.ScrollSensitivity)
	case model.StateProcess:
		m.Monitor.ProcessTable.MoveUp(m.ScrollSensitivity)
	case model.StateContainers:
		m.Monitor.Container.MoveUp(m.ScrollSensitivity)
	case model.StateNetwork:
		m.Network.NetworkTable.MoveUp(m.ScrollSensitivity)
	}
	return m, nil
}

func (m Model) handleScrollDown() (tea.Model, tea.Cmd) {
	switch m.Ui.State {
	case model.StateSystem, model.StateContainerLogs:
		m.Ui.Viewport.ScrollDown(m.ScrollSensitivity)
	case model.StateProcess:
		m.Monitor.ProcessTable.MoveDown(m.ScrollSensitivity)
	case model.StateContainers:
		m.Monitor.Container.MoveDown(m.ScrollSensitivity)
	case model.StateNetwork:
		m.Network.NetworkTable.MoveDown(m.ScrollSensitivity)
	}
	return m, nil
}
