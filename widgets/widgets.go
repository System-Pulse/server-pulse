package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/system/app"
	system "github.com/System-Pulse/server-pulse/system/app"
	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	"github.com/System-Pulse/server-pulse/system/resource"
	"github.com/System-Pulse/server-pulse/system/security"
	"github.com/System-Pulse/server-pulse/utils"
	model "github.com/System-Pulse/server-pulse/widgets/model"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
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
	case system.ContainerMsg, system.ContainerDetailsMsg, system.ContainerLogsMsg, system.ContainerOperationMsg,
		system.ExecShellMsg, system.ContainerStatsChanMsg:
		return m.handleContainerRelatedMsgs(msg)
	case security.SecurityMsg:
		return m.handleSecurityCheckMsgs(msg)
	case security.CertificateDisplayMsg:
		return m.handleCertificateDisplayMsg(msg)
	case security.SSHRootMsg:
		return m.handleSSHRootDisplayMsg(msg)
	case security.OpenedPortsMsg:
		return m.handleOpenedPortsDisplayMsg(msg)
	case security.FirewallMsg:
		return m.handleFirewallDisplayMsg(msg)
	case security.AutoBanMsg:
		return m.handleAutoBanDisplayMsg(msg)
	case system.ContainerLogsStreamMsg:
		return m.handleLogsStreamMsg(msg)
	case system.ContainerLogsStopMsg:
		return m.handleLogsStopMsg()
	case system.ContainerLogLineMsg:
		return m.handleLogLineMsg(msg)
	case model.ClearOperationMsg:
		m.LastOperationMsg = ""
	case utils.ErrMsg:
		m.Err = msg
	case utils.TickMsg:
		return m.handleTickMsg()
	case progress.FrameMsg:
		return m.handleProgressFrame(msg)
	}

	switch m.Ui.State {
	case model.StateProcess:
		if !m.Ui.SearchMode {
			m.Monitor.ProcessTable, cmd = m.Monitor.ProcessTable.Update(msg)
			cmds = append(cmds, cmd)
		}
	case model.StateContainers:
		if !m.Ui.SearchMode {
			m.Monitor.Container, cmd = m.Monitor.Container.Update(msg)
			cmds = append(cmds, cmd)
		}
	case model.StateContainerLogs:
		m.LogsViewport, cmd = m.LogsViewport.Update(msg)
		cmds = append(cmds, cmd)
	case model.StateNetwork:
		m.Network.NetworkTable, cmd = m.Network.NetworkTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.Ui.Viewport, cmd = m.Ui.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateProcessTable() tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.Ui.SearchInput.Value())

	for _, p := range m.Monitor.Processes {
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
	m.Monitor.ProcessTable.SetRows(rows)
	return nil
}

func (m *Model) updateNetworkTable() tea.Cmd {
	var rows []table.Row

	for _, iface := range m.Network.NetworkResource.Interfaces {
		statusText := "DOWN"
		if iface.Status == "up" {
			statusText = "UP"
		}

		ips := strings.Join(iface.IPs, ", ")
		if ips == "" {
			ips = "No IP"
		}

		rows = append(rows, table.Row{
			iface.Name,
			statusText,
			ips,
			utils.FormatBytes(iface.RxBytes),
			utils.FormatBytes(iface.TxBytes),
		})
	}

	m.Network.NetworkTable.SetRows(rows)

	tableHeight := min(10, len(rows)+1)
	m.Network.NetworkTable.SetHeight(tableHeight)

	return nil
}

func (m *Model) updateContainerTable(containers []app.Container) tea.Cmd {
	var rows []table.Row
	searchTerm := strings.ToLower(m.Ui.SearchInput.Value())

	for _, c := range containers {
		if searchTerm != "" && !strings.Contains(strings.ToLower(c.Image), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Name), searchTerm) &&
			!strings.Contains(strings.ToLower(c.ID), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Status), searchTerm) &&
			!strings.Contains(strings.ToLower(c.Project), searchTerm) {
			continue
		}

		statusWithIcon, health := m.getStatusWithIconForTable(c.Status, c.Health)

		rows = append(rows, table.Row{
			c.ID,
			utils.Ellipsis(c.Image, 12),
			utils.Ellipsis(c.Name, 16),
			statusWithIcon,
			health,
			c.Project,
			utils.Ellipsis(c.PortsStr, 20),
		})
	}
	m.Monitor.Container.SetRows(rows)
	return nil
}
