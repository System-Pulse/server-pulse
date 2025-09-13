package model

import (
	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
)

type MonitorModel struct {
	System               info.SystemInfo
	Cpu                  resource.CPUInfo
	Memory               resource.MemoryInfo
	Disks                []resource.DiskInfo
	Processes            []proc.ProcessInfo
	App                  *app.DockerManager
	PendingShellExec     *ShellExecRequest
	ShouldQuit           bool
	ContainerLogs        string
	ContainerLogsLoading bool
	CpuHistory           DataHistory
	MemoryHistory        DataHistory
	NetworkRxHistory     DataHistory
	NetworkTxHistory     DataHistory
	DiskReadHistory      DataHistory
	DiskWriteHistory     DataHistory
	ContainerViewState   ContainerViewState
	ContainerTabs        []string
	ContainerDetails     *app.ContainerDetails
	ContainerMenuState   ContainerMenuState
	SelectedContainer    *app.Container
	ContainerMenuItems   []ContainerMenuItem
	SelectedMenuItem     int
	CpuProgress          progress.Model
	MemProgress          progress.Model
	SwapProgress         progress.Model
	ProcessTable         table.Model
	Container            table.Model
	DiskProgress         map[string]progress.Model
}
