package app

import (
	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func (dm *DockerManager) UpdateApp() tea.Cmd {
	return func() tea.Msg {
		cont, err := dm.RefreshContainers()
		if err != nil {
			return utils.ErrMsg(err)
		}
		return ContainerMsg(cont)
	}
}

func (dm *DockerManager) RestartContainerCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.RestartContainer(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "restart",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) StartContainerCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.StartContainer(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "start",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) StopContainerCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.StopContainer(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "stop",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) PauseContainerCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.PauseContainer(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "pause",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) UnpauseContainerCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.UnpauseContainer(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "unpause",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) DeleteContainerCmd(containerID string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := dm.DeleteContainer(containerID, force)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "delete",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) ToggleContainerStateCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.ToggleContainerState(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "toggle_start",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) ToggleContainerPauseCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		err := dm.ToggleContainerPause(containerID)
		return ContainerOperationMsg{
			ContainerID: containerID,
			Operation:   "toggle_pause",
			Success:     err == nil,
			Error:       err,
		}
	}
}

func (dm *DockerManager) GetContainerLogsCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		logs, err := dm.GetContainerLogs(containerID)
		return ContainerLogsMsg{
			ContainerID: containerID,
			Logs:        logs,
			Error:       err,
		}
	}
}

func (dm *DockerManager) ExecShellCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		return ExecShellMsg{
			ContainerID: containerID,
		}
	}
}
