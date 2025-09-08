package widgets

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Update refactorisé: point d'entrée
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	case info.SystemMsg, resource.CpuMsg, resource.MemoryMsg, resource.DiskMsg, resource.NetworkMsg, proc.ProcessMsg:
		return m.handleResourceAndProcessMsgs(msg)
	case app.ContainerMsg, app.ContainerDetailsMsg, app.ContainerLogsMsg, app.ContainerOperationMsg,
		app.ExecShellMsg, app.ContainerStatsChanMsg:
		return m.handleContainerRelatedMsgs(msg)
	case ClearOperationMsg:
		m.lastOperationMsg = ""
	case utils.ErrMsg:
		m.err = msg
	case utils.TickMsg:
		return m.handleTickMsg()
	case progress.FrameMsg:
		return m.handleProgressFrame(msg)
	}

	// fallback: update interactive components (processTable / viewport)
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

// ------------------------- Handlers pour messages "système / ressources" -------------------------

func (m Model) handleResourceAndProcessMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		// cmd  tea.Cmd
	)
	switch msg := msg.(type) {
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
	}
	return m, tea.Batch(cmds...)
}

// ------------------------- Handlers pour messages liés aux conteneurs -------------------------

func (m Model) handleContainerRelatedMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case app.ContainerMsg:
		containers := []app.Container(msg)
		return m, m.updateContainerTable(containers)
	case app.ContainerDetailsMsg:
		details := app.ContainerDetails(msg)
		m.containerDetails = &details
	case app.ContainerLogsMsg:
		logsMsg := app.ContainerLogsMsg(msg)
		m.containerLogsLoading = false
		if logsMsg.Error != nil {
			m.containerLogs = fmt.Sprintf("Error loading logs: %v", logsMsg.Error)
		} else {
			m.containerLogs = logsMsg.Logs
		}
	case app.ContainerOperationMsg:
		opMsg := app.ContainerOperationMsg(msg)
		m.operationInProgress = false
		m.lastOperationMsg = utils.FormatOperationMessage(opMsg.Operation, opMsg.Success, opMsg.Error)

		var refreshCmd tea.Cmd
		if opMsg.Success {
			refreshCmd = m.app.UpdateApp()
		}
		return m, tea.Batch(refreshCmd, clearOperationMessage())
	case app.ExecShellMsg:
		m.pendingShellExec = &ShellExecRequest{ContainerID: msg.ContainerID}
		m.shouldQuit = false
		return m, tea.Quit
	case app.ContainerStatsChanMsg:
		statsMsg := app.ContainerStatsChanMsg(msg)
		go m.handleRealTimeStats(statsMsg.ContainerID, statsMsg.StatsChan)
		return m, nil
	}
	return m, nil
}

// ------------------------- Window size -------------------------

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	// Définir des tailles minimales
	minWidth := 40
	minHeight := 10

	// Si le terminal est trop petit, ne pas configurer les composants
	if msg.Width < minWidth || msg.Height < minHeight {
		return m, nil
	}

	headerHeight := lipgloss.Height(m.renderHome()) + 2
	footerHeight := lipgloss.Height(m.renderFooter())
	verticalMargin := headerHeight + footerHeight

	// S'assurer que la hauteur de la viewport est positive
	viewportHeight := msg.Height - verticalMargin
	if viewportHeight < 1 {
		viewportHeight = 1
	}

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
}

// ------------------------- Key handling (délégué) -------------------------

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Si menu de conteneurs actif -> priorité au menu
	if m.containerMenuState == ContainerMenuVisible {
		return m.handleContainerMenuKeys(msg)
	}

	// boîte de confirmation prioritaire
	if m.confirmationVisible {
		return m.handleConfirmationKeys(msg)
	}

	// mode recherche
	if m.searchMode {
		return m.handleSearchKeys(msg)
	}

	// scrolling/navigation des monitors
	if m.isMonitorActive && (m.selectedMonitor == 1 || m.selectedMonitor == 2) {
		if handled, mNew, cmd := m.handleMonitorNavigationKeys(msg); handled {
			return mNew, cmd
		}
	} else if m.activeView != -1 {
		switch msg.String() {
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		}
	}

	// Cas généraux (raccourcis, actions globales, etc.)
	return m.handleGeneralKeys(msg)
}

func (m Model) handleContainerMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedMenuItem > 0 {
			m.selectedMenuItem--
		}
		return m, nil
	case "down", "j":
		if m.selectedMenuItem < len(m.containerMenuItems)-1 {
			m.selectedMenuItem++
		}
		return m, nil
	case "enter":
		return m.executeContainerMenuAction()
	case "o":
		m.containerViewState = ContainerViewSingle
		m.containerMenuState = ContainerMenuHidden
		m.containerTab = ContainerTabGeneral
		return m, m.loadContainerDetails(m.selectedContainer.ID)
	case "l":
		m.containerViewState = ContainerViewLogs
		m.containerMenuState = ContainerMenuHidden
		m.containerLogsLoading = true
		return m, m.app.GetContainerLogsCmd(m.selectedContainer.ID, "100")
	case "r":
		m.containerMenuState = ContainerMenuHidden
		m.confirmationVisible = true
		m.confirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.selectedContainer.Name)
		m.confirmationAction = "restart"
		m.confirmationData = m.selectedContainer.ID
		return m, nil
	case "d":
		m.containerMenuState = ContainerMenuHidden
		m.confirmationVisible = true
		m.confirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.selectedContainer.Name)
		m.confirmationAction = "delete"
		m.confirmationData = m.selectedContainer.ID
		return m, nil
	case "x":
		m.containerMenuState = ContainerMenuHidden
		m.confirmationVisible = true
		m.confirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.selectedContainer.Name)
		m.confirmationAction = "remove"
		m.confirmationData = m.selectedContainer.ID
		return m, nil
	case "s":
		m.containerMenuState = ContainerMenuHidden
		m.operationInProgress = true
		return m, m.app.ToggleContainerStateCmd(m.selectedContainer.ID)
	case "p":
		m.containerMenuState = ContainerMenuHidden
		m.operationInProgress = true
		return m, m.app.ToggleContainerPauseCmd(m.selectedContainer.ID)
	case "e":
		m.containerMenuState = ContainerMenuHidden
		m.pendingShellExec = &ShellExecRequest{ContainerID: m.selectedContainer.ID}
		m.shouldQuit = false
		return m, tea.Quit
	case "t":
		m.containerMenuState = ContainerMenuHidden
		m.containerViewState = ContainerViewSingle
		m.containerTab = ContainerTabCPU
		return m, m.app.GetContainerStatsCmd(m.selectedContainer.ID)
	case "i":
		m.containerMenuState = ContainerMenuHidden
		m.containerViewState = ContainerViewSingle
		m.containerTab = ContainerTabGeneral
		return m, m.loadContainerDetails(m.selectedContainer.ID)
	case "c":
		m.containerMenuState = ContainerMenuHidden
		m.lastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	case "esc", "b":
		m.containerMenuState = ContainerMenuHidden
		m.selectedContainer = nil
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) executeContainerMenuAction() (tea.Model, tea.Cmd) {
	if m.selectedMenuItem >= len(m.containerMenuItems) {
		return m, nil
	}
	action := m.containerMenuItems[m.selectedMenuItem].Action
	switch action {
	case "open_single":
		m.containerViewState = ContainerViewSingle
		m.containerMenuState = ContainerMenuHidden
		m.containerTab = ContainerTabGeneral
		return m, m.loadContainerDetails(m.selectedContainer.ID)
	case "logs":
		m.containerViewState = ContainerViewLogs
		m.containerMenuState = ContainerMenuHidden
		m.containerLogsLoading = true
		return m, m.app.GetContainerLogsCmd(m.selectedContainer.ID, "100")
	case "restart":
		m.containerMenuState = ContainerMenuHidden
		m.confirmationVisible = true
		m.confirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.selectedContainer.Name)
		m.confirmationAction = "restart"
		m.confirmationData = m.selectedContainer.ID
		return m, nil
	case "delete":
		m.containerMenuState = ContainerMenuHidden
		m.confirmationVisible = true
		m.confirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.selectedContainer.Name)
		m.confirmationAction = "delete"
		m.confirmationData = m.selectedContainer.ID
		return m, nil
	case "remove":
		m.containerMenuState = ContainerMenuHidden
		m.confirmationVisible = true
		m.confirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.selectedContainer.Name)
		m.confirmationAction = "remove"
		m.confirmationData = m.selectedContainer.ID
		return m, nil
	case "toggle_start":
		m.containerMenuState = ContainerMenuHidden
		m.operationInProgress = true
		return m, m.app.ToggleContainerStateCmd(m.selectedContainer.ID)
	case "toggle_pause":
		m.containerMenuState = ContainerMenuHidden
		m.operationInProgress = true
		return m, m.app.ToggleContainerPauseCmd(m.selectedContainer.ID)
	case "exec":
		m.containerMenuState = ContainerMenuHidden
		m.pendingShellExec = &ShellExecRequest{ContainerID: m.selectedContainer.ID}
		m.shouldQuit = false
		return m, tea.Quit
	case "stats":
		m.containerMenuState = ContainerMenuHidden
		m.containerViewState = ContainerViewSingle
		m.containerTab = ContainerTabCPU
		return m, m.app.GetContainerStatsCmd(m.selectedContainer.ID)
	case "inspect":
		m.containerMenuState = ContainerMenuHidden
		m.containerViewState = ContainerViewSingle
		m.containerTab = ContainerTabGeneral
		return m, m.loadContainerDetails(m.selectedContainer.ID)
	case "commit":
		m.containerMenuState = ContainerMenuHidden
		m.lastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	}
	return m, nil
}

// ------------------------- Confirmation box keys -------------------------

func (m Model) handleConfirmationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.confirmationVisible = false
		switch m.confirmationAction {
		case "delete":
			if containerID, ok := m.confirmationData.(string); ok {
				m.operationInProgress = true
				m.confirmationAction = ""
				m.confirmationData = nil
				return m, m.app.DeleteContainerCmd(containerID, false)
			}
		case "remove":
			if containerID, ok := m.confirmationData.(string); ok {
				m.operationInProgress = true
				m.confirmationAction = ""
				m.confirmationData = nil
				return m, m.app.DeleteContainerCmd(containerID, true)
			}
		case "restart":
			if containerID, ok := m.confirmationData.(string); ok {
				m.operationInProgress = true
				m.confirmationAction = ""
				m.confirmationData = nil
				return m, m.app.RestartContainerCmd(containerID)
			}
		}
		return m, nil
	case "n", "N", "esc":
		m.confirmationVisible = false
		m.confirmationMessage = ""
		m.confirmationAction = ""
		m.confirmationData = nil
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
		m.searchMode = false
		m.processTable.Focus()
		m.container.Focus()
		if m.selectedMonitor == 1 {
			tcmd = m.updateProcessTable()
		} else {
			tcmd = m.app.UpdateApp()
		}
		return m, tcmd
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
}

// ------------------------- Monitor navigation (process/container) -------------------------

func (m Model) handleMonitorNavigationKeys(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedMonitor == 1 {
			m.processTable.MoveUp(1)
		} else {
			m.container.MoveUp(1)
		}
		return true, m, nil
	case "down", "j":
		if m.selectedMonitor == 1 {
			m.processTable.MoveDown(1)
		} else {
			m.container.MoveDown(1)
		}
		return true, m, nil
	case "pageup":
		if m.selectedMonitor == 1 {
			m.processTable.MoveUp(10)
		} else {
			m.container.MoveUp(10)
		}
		return true, m, nil
	case "pagedown":
		if m.selectedMonitor == 1 {
			m.processTable.MoveDown(10)
		} else {
			m.container.MoveDown(10)
		}
		return true, m, nil
	case "home":
		if m.selectedMonitor == 1 {
			m.processTable.GotoTop()
		} else {
			m.container.GotoTop()
		}
		return true, m, nil
	case "end":
		if m.selectedMonitor == 1 {
			m.processTable.GotoBottom()
		} else {
			m.container.GotoBottom()
		}
		return true, m, nil
	case "/":
		m.searchMode = true
		m.searchInput.Focus()
		return true, m, nil
	}
	return false, m, nil
}

// ------------------------- General keys & shortcuts -------------------------

func (m Model) handleGeneralKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.pendingShellExec != nil {
			return m, tea.Quit
		}
		m.shouldQuit = true
		return m, tea.Quit
	case "enter":
		if m.activeView == -1 {
			m.activeView = m.selectedTab
			if m.activeView == 0 {
				m.isMonitorActive = true
			}
		} else if m.isMonitorActive {
			if m.selectedMonitor == 2 && len(m.container.SelectedRow()) > 0 {
				selectedRow := m.container.SelectedRow()
				containerID := selectedRow[0]
				containers, err := m.app.RefreshContainers()
				if err == nil {
					for _, container := range containers {
						if container.ID == containerID {
							m.selectedContainer = &container
							m.containerMenuState = ContainerMenuVisible
							m.selectedMenuItem = 0
							break
						}
					}
				} else {
					m.selectedContainer = &app.Container{
						ID:     containerID,
						Name:   selectedRow[2],
						Image:  selectedRow[1],
						Status: selectedRow[3],
					}
					m.containerMenuState = ContainerMenuVisible
					m.selectedMenuItem = 0
				}
			}
		}
	case "b", "esc":
		if m.searchMode {
			m.searchMode = false
			m.processTable.Focus()
		} else if m.containerViewState == ContainerViewSingle {
			m.containerViewState = ContainerViewNone
			m.selectedContainer = nil
		} else if m.containerViewState == ContainerViewLogs {
			m.containerViewState = ContainerViewNone
			m.selectedContainer = nil
			m.containerLogs = ""
		} else if m.confirmationVisible {
			m.confirmationVisible = false
			m.confirmationMessage = ""
			m.confirmationAction = ""
			m.confirmationData = nil
		} else if m.containerMenuState == ContainerMenuVisible {
			m.containerMenuState = ContainerMenuHidden
			m.selectedContainer = nil
		} else if m.isMonitorActive {
			m.isMonitorActive = false
			m.activeView = -1
		} else if m.activeView != -1 {
			m.activeView = -1
		}
	case "tab", "right", "l":
		if m.containerViewState != ContainerViewSingle {
			if m.activeView == -1 {
				m.selectedTab = (m.selectedTab + 1) % len(m.tabs.DashBoard)
			} else if m.isMonitorActive {
				m.selectedMonitor = (m.selectedMonitor + 1) % len(m.tabs.Monitor)
			}
		}
	case "shift+tab", "left", "h":
		if m.containerViewState != ContainerViewSingle {
			if m.activeView == -1 {
				m.selectedTab = (m.selectedTab - 1 + len(m.tabs.DashBoard)) % len(m.tabs.DashBoard)
			} else if m.isMonitorActive {
				m.selectedMonitor = (m.selectedMonitor - 1 + len(m.tabs.Monitor)) % len(m.tabs.Monitor)
			}
		}
	case "1", "2", "3", "4":
		if m.containerViewState != ContainerViewSingle && m.activeView == -1 {
			switch msg.String() {
			case "1":
				m.selectedTab = 0
			case "2":
				m.selectedTab = 1
			case "3":
				m.selectedTab = 2
			case "4":
				m.selectedTab = 3
			}
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
	case "r":
		if m.containerViewState == ContainerViewLogs && m.selectedContainer != nil {
			m.containerLogsLoading = true
			return m, m.app.GetContainerLogsCmd(m.selectedContainer.ID, "100")
		}
	}

	if m.isMonitorActive {
		switch msg.String() {
		case "1":
			m.selectedMonitor = 0
		case "2":
			m.selectedMonitor = 1
		case "3":
			m.selectedMonitor = 2
		}
	}

	// Gestion des onglets dans la vue détaillée du conteneur
	if m.containerViewState == ContainerViewSingle {
		switch msg.String() {
		case "tab", "right", "l":
			m.containerTab = ContainerTab((int(m.containerTab) + 1) % len(m.containerTabs))
		case "shift+tab", "left", "h":
			newTab := int(m.containerTab) - 1
			if newTab < 0 {
				newTab = len(m.containerTabs) - 1
			}
			m.containerTab = ContainerTab(newTab)
		case "1":
			m.containerTab = ContainerTabGeneral
		case "2":
			m.containerTab = ContainerTabCPU
		case "3":
			m.containerTab = ContainerTabMemory
		case "4":
			m.containerTab = ContainerTabNetwork
		case "5":
			m.containerTab = ContainerTabDisk
		case "6":
			m.containerTab = ContainerTabEnv
		}
	}

	return m, nil
}

// ------------------------- Tick / périodique -------------------------

func (m Model) handleTickMsg() (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{
		tick(),
		info.UpdateSystemInfo(),
		resource.UpdateCPUInfo(),
		resource.UpdateMemoryInfo(),
		resource.UpdateDiskInfo(),
		resource.UpdateNetworkInfo(),
		proc.UpdateProcesses(),
		m.app.UpdateApp(),
	}
	// Mettre à jour les graphiques
	m.updateCharts()
	return m, tea.Batch(cmds...)
}

// ------------------------- Progress frame updates -------------------------

func (m Model) handleProgressFrame(msg progress.FrameMsg) (tea.Model, tea.Cmd) {
	var (
		progCmd tea.Cmd
		cmds    []tea.Cmd
	)

	updatedModel, progCmd := m.cpuProgress.Update(msg)
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
	return m, tea.Batch(cmds...)
}

// ------------------------- Mouse/trackpad handling -------------------------

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	mouseEvent := tea.MouseEvent(msg)

	// Éviter le scrolling trop rapide
	if time.Since(m.lastScrollTime) < 20*time.Millisecond { // ~60fps
		return m, nil
	}
	m.lastScrollTime = time.Now()

	// Gestion du scrolling avec le trackpad
	switch {
	case mouseEvent.Button == tea.MouseButtonWheelUp && mouseEvent.Action == tea.MouseActionPress:
		return m.handleScrollUp()
	case mouseEvent.Button == tea.MouseButtonWheelDown && mouseEvent.Action == tea.MouseActionPress:
		return m.handleScrollDown()
	}

	return m, nil
}

func (m Model) handleScrollUp() (tea.Model, tea.Cmd) {
	if m.isMonitorActive {
		// Scrolling dans les tables
		switch m.selectedMonitor {
		case 1:
			m.processTable.MoveUp(m.scrollSensitivity)
		case 2:
			m.container.MoveUp(m.scrollSensitivity)
		}
	} else {
		// Scrolling dans la viewport
		m.viewport.ScrollUp(m.scrollSensitivity)
	}
	return m, nil
}

func (m Model) handleScrollDown() (tea.Model, tea.Cmd) {
	if m.isMonitorActive {
		// Scrolling dans les tables
		switch m.selectedMonitor {
		case 1:
			m.processTable.MoveDown(m.scrollSensitivity)
		case 2:
			m.container.MoveDown(m.scrollSensitivity)
		}
	} else {
		// Scrolling dans la viewport
		m.viewport.ScrollDown(m.scrollSensitivity)
	}
	return m, nil
}
