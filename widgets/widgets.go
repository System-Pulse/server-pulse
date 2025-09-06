package widgets

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/utils"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
		// Gestion prioritaire du menu contextuel des conteneurs
		if m.containerMenuState == ContainerMenuVisible {
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
				// Exécuter l'action sélectionnée
				if m.selectedMenuItem < len(m.containerMenuItems) {
					action := m.containerMenuItems[m.selectedMenuItem].Action
					switch action {
					case "open_single":
						m.containerViewState = ContainerViewSingle
						m.containerMenuState = ContainerMenuHidden
						m.containerTab = ContainerTabGeneral
						// Récupérer les détails du conteneur
						return m, m.loadContainerDetails(m.selectedContainer.ID)
					case "logs":
						// TODO: Implémenter l'affichage des logs
						m.containerMenuState = ContainerMenuHidden
					case "restart":
						// TODO: Implémenter le redémarrage
						m.containerMenuState = ContainerMenuHidden
					case "delete":
						// TODO: Implémenter la suppression
						m.containerMenuState = ContainerMenuHidden
					case "toggle_start":
						// TODO: Implémenter start/stop
						m.containerMenuState = ContainerMenuHidden
					case "toggle_pause":
						// TODO: Implémenter pause/unpause
						m.containerMenuState = ContainerMenuHidden
					case "exec":
						// TODO: Implémenter exec shell
						m.containerMenuState = ContainerMenuHidden
					}
				}
				return m, nil
			case "o":
				m.containerViewState = ContainerViewSingle
				m.containerMenuState = ContainerMenuHidden
				m.containerTab = ContainerTabGeneral
				// Récupérer les détails du conteneur
				return m, m.loadContainerDetails(m.selectedContainer.ID)
			case "l":
				// TODO: Implémenter l'affichage des logs
				m.containerMenuState = ContainerMenuHidden
				return m, nil
			case "r":
				// TODO: Implémenter le redémarrage
				m.containerMenuState = ContainerMenuHidden
				return m, nil
			case "d":
				// TODO: Implémenter la suppression
				m.containerMenuState = ContainerMenuHidden
				return m, nil
			case "s":
				// TODO: Implémenter start/stop
				m.containerMenuState = ContainerMenuHidden
				return m, nil
			case "p":
				// TODO: Implémenter pause/unpause
				m.containerMenuState = ContainerMenuHidden
				return m, nil
			case "e":
				// TODO: Implémenter exec shell
				m.containerMenuState = ContainerMenuHidden
				return m, nil
			case "esc", "b":
				// Fermer le menu contextuel
				m.containerMenuState = ContainerMenuHidden
				m.selectedContainer = nil
				return m, nil
			case "q":
				return m, tea.Quit
			}
			// Si on est dans le menu, on ignore les autres touches
			return m, nil
		}

		if m.searchMode {
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
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Scrolling/navigation
		if m.isMonitorActive && (m.selectedMonitor == 1 || m.selectedMonitor == 2) {
			switch msg.String() {
			case "up", "k":
				if m.selectedMonitor == 1 {
					m.processTable.MoveUp(0)
				} else {
					m.container.MoveUp(1)
				}
			case "down", "j":
				if m.selectedMonitor == 1 {
					m.processTable.MoveDown(0)
				} else {
					m.container.MoveDown(1)
				}
			case "pageup":
				if m.selectedMonitor == 1 {
					m.processTable.MoveUp(10)
				} else {
					m.container.MoveUp(10)
				}
			case "pagedown":
				if m.selectedMonitor == 1 {
					m.processTable.MoveDown(10)
				} else {
					m.container.MoveDown(10)
				}
			case "home":
				if m.selectedMonitor == 1 {
					m.processTable.GotoTop()
				} else {
					m.container.GotoTop()
				}
			case "end":
				if m.selectedMonitor == 1 {
					m.processTable.GotoBottom()
				} else {
					m.container.GotoBottom()
				}
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
				// Si on est dans les conteneurs et qu'un conteneur est sélectionné
				if m.selectedMonitor == 2 && len(m.container.SelectedRow()) > 0 {
					// Récupérer le conteneur sélectionné
					selectedRow := m.container.SelectedRow()
					containerID := selectedRow[0]

					// Trouver le conteneur complet dans la liste
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
						// En cas d'erreur, créer un conteneur minimal à partir des données de la table
						m.selectedContainer = &app.Container{
							ID:     containerID,
							Name:   selectedRow[2], // Name column
							Image:  selectedRow[1], // Image column
							Status: selectedRow[3], // Status column
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
				// Retour de la vue détaillée vers la liste des conteneurs
				m.containerViewState = ContainerViewNone
				m.selectedContainer = nil
			} else if m.containerMenuState == ContainerMenuVisible {
				// Fermer le menu contextuel
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
		case "1":
			if m.containerViewState != ContainerViewSingle && m.activeView == -1 {
				m.selectedTab = 0
			}
		case "2":
			if m.containerViewState != ContainerViewSingle && m.activeView == -1 {
				m.selectedTab = 1
			}
		case "3":
			if m.containerViewState != ContainerViewSingle && m.activeView == -1 {
				m.selectedTab = 2
			}
		case "4":
			if m.containerViewState != ContainerViewSingle && m.activeView == -1 {
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

	case app.ContainerMsg:
		containers := []app.Container(msg)
		return m, m.updateContainerTable(containers)

	case app.ContainerDetailsMsg:
		details := app.ContainerDetails(msg)
		m.containerDetails = &details

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
			m.app.UpdateApp(),
		)
		// Mettre à jour les graphiques
		m.updateCharts()
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

func (m *model) updateContainerTable(containers []app.Container) tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.searchInput.Value())

	for _, c := range containers {
		if searchTerm != "" && !strings.Contains(strings.ToLower(c.Image), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Name), searchTerm) &&
			!strings.Contains(strings.ToLower(c.ID), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Status), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Project), searchTerm) {
			continue
		}

		rows = append(rows, table.Row{
			c.ID,
			utils.Ellipsis(c.Image, 8),
			utils.Ellipsis(c.Name, 12),
			c.Status,
			c.Project,
		})
	}
	m.container.SetRows(rows)
	return nil
}

func (m *model) loadContainerDetails(containerID string) tea.Cmd {
	return func() tea.Msg {
		details, err := m.app.GetContainerDetails(containerID)
		if err != nil {
			return utils.ErrMsg(err)
		}
		return app.ContainerDetailsMsg(*details)
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return utils.TickMsg(t)
	})
}
