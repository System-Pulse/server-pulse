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
