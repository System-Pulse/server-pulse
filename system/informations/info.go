package informations

import (
	"fmt"

	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v4/host"
)

func UpdateSystemInfo() tea.Cmd {
	return func() tea.Msg {
		hostInfo, err := host.Info()
		if err != nil {
			return utils.ErrMsg(err)
		}

		return SystemMsg{
			Hostname: hostInfo.Hostname,
			OS:       fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion),
			Kernel:   hostInfo.KernelVersion,
			Uptime:   hostInfo.Uptime,
		}
	}
}
