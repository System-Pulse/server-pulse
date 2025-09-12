package widgets

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	model "github.com/System-Pulse/server-pulse/widgets/model"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
	case model.ClearOperationMsg:
		m.LastOperationMsg = ""
	case utils.ErrMsg:
		m.Err = msg
	case utils.TickMsg:
		return m.handleTickMsg()
	case progress.FrameMsg:
		return m.handleProgressFrame(msg)
	}

	// fallback: update interactive components (processTable / viewport)
	if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 1 && !m.Ui.SearchMode {
		m.Monitor.ProcessTable, cmd = m.Monitor.ProcessTable.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
		m.Network.NetworkTable, cmd = m.Network.NetworkTable.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 2 {
		m.Monitor.Container, cmd = m.Monitor.Container.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.Ui.Viewport, cmd = m.Ui.Viewport.Update(msg)
	cmds = append(cmds, cmd)
	// else if m.Ui.ActiveView != -1 {
	// 	m.Ui.Viewport, cmd = m.Ui.Viewport.Update(msg)
	// 	cmds = append(cmds, cmd)
	// }

	return m, tea.Batch(cmds...)
}

// ------------------------- Handlers pour messages "système / ressources" -------------------------

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

// ------------------------- Handlers pour messages liés aux conteneurs -------------------------

func (m Model) handleContainerRelatedMsgs(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case app.ContainerMsg:
		containers := []app.Container(msg)
		return m, m.updateContainerTable(containers)
	case app.ContainerDetailsMsg:
		details := app.ContainerDetails(msg)
		m.Monitor.ContainerDetails = &details
	case app.ContainerLogsMsg:
		logsMsg := app.ContainerLogsMsg(msg)
		m.Monitor.ContainerLogsLoading = false
		if logsMsg.Error != nil {
			m.Monitor.ContainerLogs = fmt.Sprintf("Error loading logs: %v", logsMsg.Error)
		} else {
			m.Monitor.ContainerLogs = logsMsg.Logs
		}
	case app.ContainerOperationMsg:
		opMsg := app.ContainerOperationMsg(msg)
		m.OperationInProgress = false
		m.LastOperationMsg = utils.FormatOperationMessage(opMsg.Operation, opMsg.Success, opMsg.Error)

		var refreshCmd tea.Cmd
		if opMsg.Success {
			refreshCmd = m.Monitor.App.UpdateApp()
		}
		return m, tea.Batch(refreshCmd, clearOperationMessage())
	case app.ExecShellMsg:
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: msg.ContainerID}
		m.Monitor.ShouldQuit = false
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
	m.Ui.Width = msg.Width
	m.Ui.Height = msg.Height

	minWidth := 40
	minHeight := 10

	if msg.Width < minWidth || msg.Height < minHeight {
		return m, nil
	}

	headerHeight := lipgloss.Height(m.renderHome()) + 2
	footerHeight := lipgloss.Height(m.renderFooter())
	verticalMargin := headerHeight + footerHeight

	// viewportHeight := max(msg.Height - verticalMargin, 1)

	m.Ui.Viewport.Width = msg.Width
	m.Ui.Viewport.Height = msg.Height - verticalMargin
	m.Monitor.ProcessTable.SetWidth(msg.Width)
	m.Monitor.ProcessTable.SetHeight(m.Ui.Viewport.Height - 1)

	progWidth := min(max(msg.Width/3, 20), progressBarWidth)
	m.Monitor.CpuProgress.Width = progWidth
	m.Monitor.MemProgress.Width = progWidth
	m.Monitor.SwapProgress.Width = progWidth
	for k, p := range m.Monitor.DiskProgress {
		p.Width = progWidth
		m.Monitor.DiskProgress[k] = p
	}

	if !m.Ui.Ready {
		m.Ui.Ready = true
	}
	return m, nil
}

// ------------------------- Key handling (délégué) -------------------------

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.Monitor.ContainerMenuState == ContainerMenuVisible {
		return m.handleContainerMenuKeys(msg)
	}

	if m.ConfirmationVisible {
		return m.handleConfirmationKeys(msg)
	}

	if m.Ui.SearchMode {
		return m.handleSearchKeys(msg)
	}

	// Gestion spécifique pour l'onglet Network
	if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
		// D'abord, vérifier les touches de navigation réseau
		handled, updatedModel, cmd := m.handleNetwork(msg)
		if handled {
			return updatedModel, cmd
		}
		// Ensuite, vérifier les touches de navigation du tableau
		if handled, mNew, cmd := m.handleMonitorNavigationKeys(msg); handled {
			return mNew, cmd
		}
		// Enfin, vérifier les touches générales
		return m.handleGeneralKeys(msg)
	}

	if m.Ui.IsMonitorActive && (m.Ui.SelectedMonitor == 1 || m.Ui.SelectedMonitor == 2) {
		if handled, mNew, cmd := m.handleMonitorNavigationKeys(msg); handled {
			return mNew, cmd
		}
	} else if m.Ui.ActiveView != -1 && !m.Ui.IsNetworkActive {
		switch msg.String() {
		case "up", "k":
			m.Ui.Viewport.ScrollUp(1)
		case "down", "j":
			m.Ui.Viewport.ScrollDown(1)
		}
	}

	return m.handleGeneralKeys(msg)
}

func (m Model) handleNetwork(msg tea.KeyMsg) (bool, Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "right", "l":
		newTab := int(m.Network.SelectedItem) + 1
		if newTab >= len(m.Network.Nav) {
			newTab = 0
		}
		m.Network.SelectedItem = model.ContainerTab(newTab)
		return true, m, nil
	case "shift+tab", "left", "h":
		newTab := int(m.Network.SelectedItem) - 1
		if newTab < 0 {
			newTab = len(m.Network.Nav) - 1
		}
		m.Network.SelectedItem = model.ContainerTab(newTab)
		return true, m, nil
	case "1", "2", "3", "4":
		// Raccourcis numériques pour les onglets réseau
		switch msg.String() {
		case "1":
			m.Network.SelectedItem = model.NetworkTabInterface
		case "2":
			m.Network.SelectedItem = model.NetworkTabConnectivity
		case "3":
			m.Network.SelectedItem = model.NetworkTabConfiguration
		case "4":
			m.Network.SelectedItem = model.NetworkTabProtocol
		}
		return true, m, nil
	case "b", "esc":
		// Retour en arrière
		if m.Ui.IsNetworkActive {
			m.Ui.IsNetworkActive = false
			m.Ui.ActiveView = -1
			return true, m, nil
		}
	case "q", "ctrl+c":
		// Quitter
		m.Monitor.ShouldQuit = true
		return true, m, tea.Quit
	}

	// Si aucune touche spécifique n'est traitée
	return false, m, nil
}

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
		m.Monitor.ContainerViewState = ContainerViewSingle
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "l":
		m.Monitor.ContainerViewState = ContainerViewLogs
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.ContainerLogsLoading = true
		return m, m.Monitor.App.GetContainerLogsCmd(m.Monitor.SelectedContainer.ID, "100")
	case "r":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "restart"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "d":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "delete"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "x":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "remove"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "s":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerStateCmd(m.Monitor.SelectedContainer.ID)
	case "p":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerPauseCmd(m.Monitor.SelectedContainer.ID)
	case "e":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: m.Monitor.SelectedContainer.ID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case "t":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.ContainerViewState = ContainerViewSingle
		m.ContainerTab = model.ContainerTabCPU
		return m, m.Monitor.App.GetContainerStatsCmd(m.Monitor.SelectedContainer.ID)
	case "i":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.ContainerViewState = ContainerViewSingle
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "c":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.LastOperationMsg = "Commit functionality not yet implemented"
		return m, nil
	case "esc", "b":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.SelectedContainer = nil
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) executeContainerMenuAction() (tea.Model, tea.Cmd) {
	if m.Monitor.SelectedMenuItem >= len(m.Monitor.ContainerMenuItems) {
		return m, nil
	}
	action := m.Monitor.ContainerMenuItems[m.Monitor.SelectedMenuItem].Action
	switch action {
	case "open_single":
		m.Monitor.ContainerViewState = ContainerViewSingle
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "logs":
		m.Monitor.ContainerViewState = ContainerViewLogs
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.ContainerLogsLoading = true
		return m, m.Monitor.App.GetContainerLogsCmd(m.Monitor.SelectedContainer.ID, "100")
	case "restart":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Restart container '%s'?\nThis will stop and start the container.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "restart"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "delete":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Delete container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "delete"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "remove":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.ConfirmationVisible = true
		m.ConfirmationMessage = fmt.Sprintf("Force remove container '%s'?\nThis action cannot be undone.", m.Monitor.SelectedContainer.Name)
		m.ConfirmationAction = "remove"
		m.ConfirmationData = m.Monitor.SelectedContainer.ID
		return m, nil
	case "toggle_start":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerStateCmd(m.Monitor.SelectedContainer.ID)
	case "toggle_pause":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.OperationInProgress = true
		return m, m.Monitor.App.ToggleContainerPauseCmd(m.Monitor.SelectedContainer.ID)
	case "exec":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.PendingShellExec = &model.ShellExecRequest{ContainerID: m.Monitor.SelectedContainer.ID}
		m.Monitor.ShouldQuit = false
		return m, tea.Quit
	case "stats":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.ContainerViewState = ContainerViewSingle
		m.ContainerTab = model.ContainerTabCPU
		return m, m.Monitor.App.GetContainerStatsCmd(m.Monitor.SelectedContainer.ID)
	case "inspect":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
		m.Monitor.ContainerViewState = ContainerViewSingle
		m.ContainerTab = model.ContainerTabGeneral
		return m, m.loadContainerDetails(m.Monitor.SelectedContainer.ID)
	case "commit":
		m.Monitor.ContainerMenuState = ContainerMenuHidden
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

// ------------------------- Monitor navigation (process/container) -------------------------

func (m Model) handleMonitorNavigationKeys(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
			m.Network.NetworkTable.MoveUp(1)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 1 {
			m.Monitor.ProcessTable.MoveUp(1)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 2 {
			m.Monitor.Container.MoveUp(1)
			return true, m, nil
		}
	case "down", "j":
		if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
			m.Network.NetworkTable.MoveDown(1)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 1 {
			m.Monitor.ProcessTable.MoveDown(1)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 2 {
			m.Monitor.Container.MoveDown(1)
			return true, m, nil
		}
	case "pageup":
		if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
			m.Network.NetworkTable.MoveUp(10)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 1 {
			m.Monitor.ProcessTable.MoveUp(10)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 2 {
			m.Monitor.Container.MoveUp(10)
			return true, m, nil
		}
	case "pagedown":
		if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
			m.Network.NetworkTable.MoveDown(10)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 1 {
			m.Monitor.ProcessTable.MoveDown(10)
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 2 {
			m.Monitor.Container.MoveDown(10)
			return true, m, nil
		}
	case "home":
		if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
			m.Network.NetworkTable.GotoTop()
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 1 {
			m.Monitor.ProcessTable.GotoTop()
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 2 {
			m.Monitor.Container.GotoTop()
			return true, m, nil
		}
	case "end":
		if m.Ui.ActiveView == 2 && m.Ui.IsNetworkActive {
			m.Network.NetworkTable.GotoBottom()
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 1 {
			m.Monitor.ProcessTable.GotoBottom()
			return true, m, nil
		} else if m.Ui.SelectedMonitor == 2 {
			m.Monitor.Container.GotoBottom()
			return true, m, nil
		}
	case "/":
		m.Ui.SearchMode = true
		m.Ui.SearchInput.Focus()
		return true, m, nil
	}
	return false, m, nil
}

// ------------------------- General keys & shortcuts -------------------------

func (m Model) handleGeneralKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.Monitor.PendingShellExec != nil {
			return m, tea.Quit
		}
		m.Monitor.ShouldQuit = true
		return m, tea.Quit
	case "enter":
		if m.Ui.ActiveView == -1 {
			m.Ui.ActiveView = m.Ui.SelectedTab
			switch m.Ui.ActiveView {
			case 0:
				m.Ui.IsMonitorActive = true
			case 2:
				m.Ui.IsNetworkActive = true
			}
		} else if m.Ui.IsMonitorActive {
			if m.Ui.SelectedMonitor == 2 && len(m.Monitor.Container.SelectedRow()) > 0 {
				selectedRow := m.Monitor.Container.SelectedRow()
				containerID := selectedRow[0]
				containers, err := m.Monitor.App.RefreshContainers()
				if err == nil {
					for _, container := range containers {
						if container.ID == containerID {
							m.Monitor.SelectedContainer = &container
							m.Monitor.ContainerMenuState = ContainerMenuVisible
							m.Monitor.SelectedMenuItem = 0
							break
						}
					}
				} else {
					m.Monitor.SelectedContainer = &app.Container{
						ID:     containerID,
						Name:   selectedRow[2],
						Image:  selectedRow[1],
						Status: selectedRow[3],
					}
					m.Monitor.ContainerMenuState = ContainerMenuVisible
					m.Monitor.SelectedMenuItem = 0
				}
			}
		}
	case "b", "esc":
		if m.Ui.SearchMode {
			m.Ui.SearchMode = false
			m.Monitor.ProcessTable.Focus()
		} else if m.Monitor.ContainerViewState == ContainerViewSingle {
			m.Monitor.ContainerViewState = ContainerViewNone
			m.Monitor.SelectedContainer = nil
		} else if m.Monitor.ContainerViewState == ContainerViewLogs {
			m.Monitor.ContainerViewState = ContainerViewNone
			m.Monitor.SelectedContainer = nil
			m.Monitor.ContainerLogs = ""
		} else if m.ConfirmationVisible {
			m.ConfirmationVisible = false
			m.ConfirmationMessage = ""
			m.ConfirmationAction = ""
			m.ConfirmationData = nil
		} else if m.Monitor.ContainerMenuState == ContainerMenuVisible {
			m.Monitor.ContainerMenuState = ContainerMenuHidden
			m.Monitor.SelectedContainer = nil
		} else if m.Ui.IsMonitorActive {
			m.Ui.IsMonitorActive = false
			m.Ui.ActiveView = -1
		} else if m.Ui.ActiveView != -1 {
			m.Ui.ActiveView = -1
		} else if m.Ui.IsNetworkActive {
			m.Ui.IsNetworkActive = false
			m.Ui.ActiveView = -1
		}
	case "tab", "right", "l":
		if m.Monitor.ContainerViewState != ContainerViewSingle {
			if m.Ui.ActiveView == -1 {
				m.Ui.SelectedTab = (m.Ui.SelectedTab + 1) % len(m.Ui.Tabs.DashBoard)
			} else if m.Ui.IsMonitorActive {
				m.Ui.SelectedMonitor = (m.Ui.SelectedMonitor + 1) % len(m.Ui.Tabs.Monitor)
			}
		}
	case "shift+tab", "left", "h":
		if m.Monitor.ContainerViewState != ContainerViewSingle {
			if m.Ui.ActiveView == -1 {
				m.Ui.SelectedTab = (m.Ui.SelectedTab - 1 + len(m.Ui.Tabs.DashBoard)) % len(m.Ui.Tabs.DashBoard)
			} else if m.Ui.IsMonitorActive {
				m.Ui.SelectedMonitor = (m.Ui.SelectedMonitor - 1 + len(m.Ui.Tabs.Monitor)) % len(m.Ui.Tabs.Monitor)
			}
		}
	case "1", "2", "3", "4":
		if m.Monitor.ContainerViewState != ContainerViewSingle && m.Ui.ActiveView == -1 {
			switch msg.String() {
			case "1":
				m.Ui.SelectedTab = 0
			case "2":
				m.Ui.SelectedTab = 1
			case "3":
				m.Ui.SelectedTab = 2
			case "4":
				m.Ui.SelectedTab = 3
			}
		}
	case "k":
		if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 1 && len(m.Monitor.ProcessTable.SelectedRow()) > 0 {
			selectedPID := m.Monitor.ProcessTable.SelectedRow()[0]
			pid, _ := strconv.Atoi(selectedPID)
			process, _ := os.FindProcess(pid)
			if process != nil {
				_ = process.Kill()
			}
			return m, proc.UpdateProcesses()
		}
	case "s":
		if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 1 {
			sort.Slice(m.Monitor.Processes, func(i, j int) bool {
				return m.Monitor.Processes[i].CPU > m.Monitor.Processes[j].CPU
			})
			return m, m.updateProcessTable()
		}
	case "m":
		if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 1 {
			sort.Slice(m.Monitor.Processes, func(i, j int) bool {
				return m.Monitor.Processes[i].Mem > m.Monitor.Processes[j].Mem
			})
			return m, m.updateProcessTable()
		}
	case "r":
		if m.Monitor.ContainerViewState == ContainerViewLogs && m.Monitor.SelectedContainer != nil {
			m.Monitor.ContainerLogsLoading = true
			return m, m.Monitor.App.GetContainerLogsCmd(m.Monitor.SelectedContainer.ID, "100")
		}
	}

	if m.Ui.IsMonitorActive &&
		m.Monitor.ContainerViewState != ContainerViewSingle &&
		!m.Ui.IsNetworkActive {
		switch msg.String() {
		case "1":
			m.Ui.SelectedMonitor = 0
		case "2":
			m.Ui.SelectedMonitor = 1
		case "3":
			m.Ui.SelectedMonitor = 2
		}
	}

	if m.Monitor.ContainerViewState == ContainerViewSingle {
		switch msg.String() {
		case "tab", "right", "l":
			m.ContainerTab = model.ContainerTab((int(m.ContainerTab) + 1) % len(m.Monitor.ContainerTabs))
		case "shift+tab", "left", "h":
			newTab := int(m.ContainerTab) - 1
			if newTab < 0 {
				newTab = len(m.Monitor.ContainerTabs) - 1
			}
			m.ContainerTab = model.ContainerTab(newTab)
		case "1":
			m.ContainerTab = model.ContainerTabGeneral
		case "2":
			m.ContainerTab = model.ContainerTabCPU
		case "3":
			m.ContainerTab = model.ContainerTabMemory
		case "4":
			m.ContainerTab = model.ContainerTabNetwork
		case "5":
			m.ContainerTab = model.ContainerTabDisk
		case "6":
			m.ContainerTab = model.ContainerTabEnv
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
		m.Monitor.App.UpdateApp(),
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
	if m.Ui.IsMonitorActive {
		switch m.Ui.SelectedMonitor {
		case 1:
			m.Monitor.ProcessTable.MoveUp(m.ScrollSensitivity)
		case 2:
			m.Monitor.Container.MoveUp(m.ScrollSensitivity)
		}
	} else if m.Ui.ActiveView == 2 {
		m.Network.NetworkTable.MoveUp(m.ScrollSensitivity)
	} else {
		m.Ui.Viewport.ScrollUp(m.ScrollSensitivity)
	}
	return m, nil
}

func (m Model) handleScrollDown() (tea.Model, tea.Cmd) {
	if m.Ui.IsMonitorActive {
		switch m.Ui.SelectedMonitor {
		case 1:
			m.Monitor.ProcessTable.MoveDown(m.ScrollSensitivity)
		case 2:
			m.Monitor.Container.MoveDown(m.ScrollSensitivity)
		}
	} else if m.Ui.ActiveView == 2 {
		m.Network.NetworkTable.MoveDown(m.ScrollSensitivity)
	} else {
		m.Ui.Viewport.ScrollDown(m.ScrollSensitivity)
	}
	return m, nil
}
