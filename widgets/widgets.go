package widgets

import (
	"context"
	"fmt"
	"io"
	"log"
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

	"github.com/moby/moby/api/types/container"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
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
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		// Gestion des touches pour la vue unique du conteneur
		if m.containerSingleView.Visible {
			switch msg.String() {
			case "esc", "b":
				m.containerSingleView.Visible = false
				return m, nil
			case "tab", "right", "l":
				m.containerSingleView.ActiveTab = (m.containerSingleView.ActiveTab + 1) % len(m.containerSingleView.Tabs)
				return m, nil
			case "shift+tab", "left", "h":
				m.containerSingleView.ActiveTab = (m.containerSingleView.ActiveTab - 1 + len(m.containerSingleView.Tabs)) % len(m.containerSingleView.Tabs)
				return m, nil
			case "1":
				m.containerSingleView.ActiveTab = 0
				return m, nil
			case "2":
				m.containerSingleView.ActiveTab = 1
				return m, nil
			case "3":
				m.containerSingleView.ActiveTab = 2
				return m, nil
			case "4":
				m.containerSingleView.ActiveTab = 3
				return m, nil
			case "5":
				m.containerSingleView.ActiveTab = 4
				return m, nil
			case "6":
				m.containerSingleView.ActiveTab = 5
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		if m.containerMenu.Visible {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "b", "esc":
				m.containerMenu.Visible = false
				m.container.Focus()
				return m, nil
			case "up", "k":
				if m.containerMenu.Selected > 0 {
					m.containerMenu.Selected--
				}
				return m, nil
			case "down", "j":
				options := m.getMenuOptions()
				if m.containerMenu.Selected < len(options)-1 {
					m.containerMenu.Selected++
				}
				return m, nil
			case "enter":
				return m, m.executeContainerAction()
			case "o", "l", "r", "d", "s", "p", "u", "e":
				return m, m.handleContainerShortcut(msg.String())
			}
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
			case "space":
				if m.selectedMonitor == 2 && len(m.container.SelectedRow()) > 0 {
					containerID := m.container.SelectedRow()[0]
					if containerData := m.findContainerByID(containerID); containerData != nil {
						m.showContainerMenu(*containerData)
					}
				}
				return m, nil
			case "enter":
				if m.selectedMonitor == 2 && len(m.container.SelectedRow()) > 0 {
					containerID := m.container.SelectedRow()[0]
					if containerData := m.findContainerByID(containerID); containerData != nil {
						m.showContainerMenu(*containerData)
					}
				}
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
				// NOTHING
			}
		case "b", "esc":
			if m.searchMode {
				m.searchMode = false
				m.processTable.Focus()
			} else if m.isMonitorActive {
				m.isMonitorActive = false
				m.activeView = -1
			} else if m.activeView != -1 {
				m.activeView = -1
			}
		case "tab", "right", "l":
			if m.activeView == -1 {
				m.selectedTab = (m.selectedTab + 1) % len(m.tabs.DashBoard)
			} else if m.isMonitorActive {
				m.selectedMonitor = (m.selectedMonitor + 1) % len(m.tabs.Monitor)
			}
		case "shift+tab", "left", "h":
			if m.activeView == -1 {
				m.selectedTab = (m.selectedTab - 1 + len(m.tabs.DashBoard)) % len(m.tabs.DashBoard)
			} else if m.isMonitorActive {
				m.selectedMonitor = (m.selectedMonitor - 1 + len(m.tabs.Monitor)) % len(m.tabs.Monitor)
			}
		case "1":
			if m.activeView == -1 {
				m.selectedTab = 0
			}
		case "2":
			if m.activeView == -1 {
				m.selectedTab = 1
			}
		case "3":
			if m.activeView == -1 {
				m.selectedTab = 2
			}
		case "4":
			if m.activeView == -1 {
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

	case utils.ErrMsg:
		m.err = msg
	case utils.InfoMsg:
		// Handle info messages if needed
	case LogsMsg:
		// Handle logs messages if needed
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

		// Si la vue unique du conteneur est active, mettre à jour ses statistiques
		if m.containerSingleView.Visible {
			cmds = append(cmds, m.app.GetContainerStats(m.containerSingleView.Container.ID))
		}
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
	case ContainerActionMsg:
		m.containerMenu.Visible = false
		m.container.Focus()

		switch msg.Action {
		case "open single view":
			return m, m.openContainerSingleView(msg.Container)
		case "view container logs":
			return m, m.viewContainerLogs(msg.Container)
		case "restart container":
			return m, m.restartContainer(msg.Container)
		case "delete container":
			return m, m.deleteContainer(msg.Container)
		case "stop container", "start container":
			return m, m.toggleContainerState(msg.Container)
		case "pause container", "unpause container":
			return m, m.togglePauseContainer(msg.Container)
		case "exec shell":
			return m, m.execContainerShell(msg.Container)
		}
	case OpenSingleViewMsg:
		m.containerSingleView.Visible = true
		m.containerSingleView.Container = msg.Container
		m.containerSingleView.ActiveTab = 0
		m.containerSingleView.Stats = msg.Stats
		m.containerSingleView.EnvVars = msg.EnvVars
		// Commencer à collecter les statistiques
		return m, m.app.GetContainerStats(msg.Container.ID)
	case app.ContainerStatsMsg:
		if m.containerSingleView.Visible && m.containerSingleView.Container.ID == msg.ContainerID {
			// Mettre à jour les statistiques actuelles
			m.containerSingleView.Stats.CPUUsage = msg.CPUPercent
			m.containerSingleView.Stats.MemUsage = msg.MemPercent
			m.containerSingleView.Stats.MemLimit = msg.MemLimit
			m.containerSingleView.Stats.NetRX = msg.NetRX
			m.containerSingleView.Stats.NetTX = msg.NetTX
			m.containerSingleView.Stats.DiskUsage = msg.DiskUsage

			// Ajouter aux données historiques (garder seulement les 50 dernières valeurs)
			m.containerSingleView.CPUData = append(m.containerSingleView.CPUData[1:], msg.CPUPercent)
			m.containerSingleView.MemoryData = append(m.containerSingleView.MemoryData[1:], msg.MemPercent)

			// Calculer les vitesses réseau (approximation)
			m.containerSingleView.NetworkRX = append(m.containerSingleView.NetworkRX[1:], float64(msg.NetRX))
			m.containerSingleView.NetworkTX = append(m.containerSingleView.NetworkTX[1:], float64(msg.NetTX))
			m.containerSingleView.DiskData = append(m.containerSingleView.DiskData[1:], float64(msg.DiskUsage))

			// Continuer à collecter les statistiques si la vue est encore active
			if m.containerSingleView.Visible {
				return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
					cmd := m.app.GetContainerStats(m.containerSingleView.Container.ID)
					return cmd()
				})
			}
		}
		return m, nil
	}

	if m.isMonitorActive && m.selectedMonitor == 1 {
		if !m.searchMode {
			m.processTable, cmd = m.processTable.Update(msg)
			cmds = append(cmds, cmd)
		}
	} else if m.isMonitorActive && m.selectedMonitor == 2 {
		if !m.searchMode {
			m.container, cmd = m.container.Update(msg)
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

func (m *model) getMenuOptions() []string {
	options := []string{
		"open single view",
		"view container logs",
		"restart container",
		"delete container",
	}

	if strings.Contains(strings.ToLower(m.containerMenu.Container.Status), "running") {
		options = append(options, "stop container", "pause container")
	} else if strings.Contains(strings.ToLower(m.containerMenu.Container.Status), "stopped") {
		options = append(options, "start container")
	} else if strings.Contains(strings.ToLower(m.containerMenu.Container.Status), "paused") {
		options = append(options, "unpause container")
	}

	options = append(options, "exec shell")
	return options
}

func (m *model) showContainerMenu(container app.Container) {
	m.containerMenu = ContainerMenuState{
		Visible:   true,
		Container: container,
		Selected:  0,
		X:         m.width / 2,  //2 - 20,
		Y:         m.height / 3, //2 - 10,
	}
	m.container.Blur()
}

func (m *model) findContainerByID(id string) *app.Container {
	// Cette méthode devra être améliorée pour récupérer le conteneur complet
	// depuis les données stockées ou via l'API Docker
	rows := m.container.Rows()
	for _, row := range rows {
		if len(row) > 0 && row[0] == id {
			return &app.Container{
				ID:     id,
				Name:   row[2],
				Status: row[3],
				Image:  row[1],
			}
		}
	}
	return nil
}

func (m *model) handleContainerShortcut(key string) tea.Cmd {
	actions := map[string]string{
		"o": "open single view",
		"l": "view container logs",
		"r": "restart container",
		"d": "delete container",
		"s": "stop/start container",
		"p": "pause container",
		"u": "unpause container",
		"e": "exec shell",
	}

	if action, ok := actions[key]; ok {
		return func() tea.Msg {
			return ContainerActionMsg{
				Action:    action,
				Container: m.containerMenu.Container,
			}
		}
	}
	return nil
}

func (m *model) executeContainerAction() tea.Cmd {
	options := m.getMenuOptions()
	if m.containerMenu.Selected < len(options) {
		action := options[m.containerMenu.Selected]
		return func() tea.Msg {
			return ContainerActionMsg{
				Action:    action,
				Container: m.containerMenu.Container,
			}
		}
	}
	return nil
}

func (m *model) openContainerSingleView(container app.Container) tea.Cmd {
	return func() tea.Msg {
		// Récupérer les statistiques du conteneur
		ctx := context.Background()

		// Inspecter le conteneur pour obtenir les détails
		containerJSON, err := m.app.Cli.ContainerInspect(ctx, container.ID)
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to inspect container: %w", err))
		}

		// Remplir les informations de base
		stats := ContainerStats{
			ID:      container.ID,
			Name:    container.Name,
			Image:   container.Image,
			State:   containerJSON.State.Status,
			Created: containerJSON.Created,
			Health:  getHealthStatus(containerJSON.State),
		}

		// Extraire les ports
		var ports []string
		for _, port := range container.Ports {
			if port.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
			}
		}
		stats.Ports = strings.Join(ports, ", ")

		// Extraire les IPs
		if containerJSON.NetworkSettings != nil {
			for _, network := range containerJSON.NetworkSettings.Networks {
				if network.IPAddress != "" {
					stats.IPs = append(stats.IPs, network.IPAddress)
				}
			}
		}

		// Calculer l'uptime
		if containerJSON.State.StartedAt != "" {
			startTime, err := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
			if err == nil {
				uptime := time.Since(startTime)
				stats.Uptime = utils.FormatUptime(uint64(uptime.Seconds()))
			}
		}

		// Récupérer les variables d'environnement
		envVars := make(map[string]string)
		for _, env := range containerJSON.Config.Env {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}

		return OpenSingleViewMsg{
			Container: container,
			Stats:     stats,
			EnvVars:   envVars,
		}
	}
}

func getHealthStatus(state *container.State) string {
	if state.Health == nil {
		return "No health check"
	}
	return state.Health.Status
}

func (m *model) viewContainerLogs(cont app.Container) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "100",
		}
		reader, err := m.app.Cli.ContainerLogs(ctx, cont.ID, options)
		if err != nil {
			return utils.ErrMsg(err)
		}
		defer reader.Close()

		logs, err := io.ReadAll(reader)
		if err != nil {
			return utils.ErrMsg(err)
		}
		return LogsMsg{Logs: string(logs), Container: cont}
	}
}

func (m *model) restartContainer(cont app.Container) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		timeout := 10
		options := container.StopOptions{Timeout: &timeout}

		err := m.app.Cli.ContainerRestart(ctx, cont.ID, options)
		if err != nil {
			return utils.ErrMsg(err)
		}
		return utils.InfoMsg("Container restarted: " + cont.Name)
	}
}

func (m *model) deleteContainer(cont app.Container) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		options := container.RemoveOptions{Force: true}

		err := m.app.Cli.ContainerRemove(ctx, cont.ID, options)
		if err != nil {
			return utils.ErrMsg(err)
		}
		return utils.InfoMsg("Container deleted: " + cont.Name)
	}
}

func (m *model) toggleContainerState(cont app.Container) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if strings.Contains(strings.ToLower(cont.Status), "running") {
			timeout := 10
			options := container.StopOptions{Timeout: &timeout}
			err := m.app.Cli.ContainerStop(ctx, cont.ID, options)
			if err != nil {
				return utils.ErrMsg(err)
			}
			return utils.InfoMsg("Container stopped: " + cont.Name)
		} else {
			options := container.StartOptions{}
			err := m.app.Cli.ContainerStart(ctx, cont.ID, options)
			if err != nil {
				return utils.ErrMsg(err)
			}
			return utils.InfoMsg("Container started: " + cont.Name)
		}
	}
}

func (m *model) togglePauseContainer(cont app.Container) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if strings.Contains(strings.ToLower(cont.Status), "paused") {
			err := m.app.Cli.ContainerUnpause(ctx, cont.ID)
			if err != nil {
				return utils.ErrMsg(err)
			}
			return utils.InfoMsg("Container unpaused: " + cont.Name)
		} else {
			err := m.app.Cli.ContainerPause(ctx, cont.ID)
			if err != nil {
				return utils.ErrMsg(err)
			}
			return utils.InfoMsg("Container paused: " + cont.Name)
		}
	}
}

func (m *model) execContainerShell(cont app.Container) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Créer une exécution interactive dans le conteneur
		execConfig := container.ExecOptions{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Cmd:          []string{"/bin/sh", "-c", "printf '\\e[0m\\e[?25h' && clear && eval `grep ^$(id -un): /etc/passwd | cut -d : -f 7-`"},
		}

		// Créer l'exécution
		execID, err := m.app.Cli.ContainerExecCreate(ctx, cont.ID, execConfig)
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to create exec: %w", err))
		}

		// Démarrer l'exécution avec les streams attachés
		hijackedResp, err := m.app.Cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{
			Detach: false, // Ne pas se détacher
			Tty:    true,  // Utiliser TTY
		})
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to attach exec: %w", err))
		}
		defer hijackedResp.Close()

		// Configurer le terminal pour l'interaction
		oldState, err := setRawTerminal()
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to set raw terminal: %w", err))
		}
		defer restoreTerminal(oldState)

		// Redimensionner le TTY du conteneur
		if err := m.app.Cli.ContainerExecResize(ctx, execID.ID, container.ResizeOptions{
			Height: uint(m.height),
			Width:  uint(m.width),
		}); err != nil {
			log.Printf("Warning: failed to resize exec: %v", err)
		}

		// Copier les streams entre le terminal et le conteneur
		go func() {
			io.Copy(hijackedResp.Conn, os.Stdin)
		}()
		io.Copy(os.Stdout, hijackedResp.Reader)

		// Inspecter l'exécution pour vérifier le code de sortie
		execInspect, err := m.app.Cli.ContainerExecInspect(ctx, execID.ID)
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to inspect exec: %w", err))
		}

		if execInspect.ExitCode != 0 {
			return utils.InfoMsg(fmt.Sprintf("Shell exited with code %d", execInspect.ExitCode))
		}

		return utils.InfoMsg("Shell session completed for container: " + cont.Name)
	}
}

/*
func (m *model) execContainerShellDirect(cont app.Container) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        execConfig := container.ExecOptions{
            AttachStdin:  true,
            AttachStdout: true,
            AttachStderr: true,
            Tty:          true,
            Cmd:          []string{"/bin/sh", "-c", "printf '\\e[0m\\e[?25h' && clear && eval `grep ^$(id -un): /etc/passwd | cut -d : -f 7-`"},
        }

        execID, err := m.app.Cli.ContainerExecCreate(ctx, cont.ID, execConfig)
        if err != nil {
            return utils.ErrMsg(fmt.Errorf("failed to create exec: %w", err))
        }

        // CORRECTION: Utiliser la bonne structure
        hijackedResp, err := m.app.Cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{
            Detach: false,
            Tty:    true,
        })
        if err != nil {
            return utils.ErrMsg(fmt.Errorf("failed to attach exec: %w", err))
        }
        defer hijackedResp.Close()

        oldState, err := setRawTerminal()
        if err != nil {
            return utils.ErrMsg(fmt.Errorf("failed to set raw terminal: %w", err))
        }
        defer restoreTerminal(oldState)

        // Redimensionner
        m.app.Cli.ContainerExecResize(ctx, execID.ID, container.ResizeOptions{
            Height: uint(m.height),
            Width:  uint(m.width),
        })

        // Copier les streams
        go func() {
            io.Copy(hijackedResp.Conn, os.Stdin)
        }()
        io.Copy(os.Stdout, hijackedResp.Reader)

        return utils.InfoMsg("Shell session completed")
    }
}
*/

// Ajouter ces fonctions d'aide pour la gestion du terminal
func setRawTerminal() (*term.State, error) {
	fd := os.Stdin.Fd()
	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("stdin is not a terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to set raw terminal: %w", err)
	}

	return oldState, nil
}

func restoreTerminal(oldState *term.State) {
	fd := os.Stdin.Fd()
	if term.IsTerminal(fd) && oldState != nil {
		term.Restore(fd, oldState)
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return utils.TickMsg(t)
	})
}
