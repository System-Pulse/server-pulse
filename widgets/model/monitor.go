package model

import (
	"context"

	info "github.com/System-Pulse/server-pulse/system/informations"
	proc "github.com/System-Pulse/server-pulse/system/process"
	resource "github.com/System-Pulse/server-pulse/system/resource"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
)

type ContainerLogsPagination struct {
	CurrentPage int
	TotalPages  int
	PageSize    int
	Lines       []string
}

type ProcessSortField int

const (
	ProcessSortByCPU ProcessSortField = iota
	ProcessSortByMem
)

type MonitorModel struct {
	System                  info.SystemInfo
	Cpu                     resource.CPUInfo
	Memory                  resource.MemoryInfo
	Disks                   []resource.DiskInfo
	Processes               []proc.ProcessInfo
	ProcessSort             ProcessSortField
	App                     *app.DockerManager
	PendingShellExec        *ShellExecRequest
	ShouldQuit              bool
	ContainerLogsPagination ContainerLogsPagination

	ContainerLogsStreaming bool
	ContainerLogsChan      chan string
	LogsCancelFunc         context.CancelFunc
	StatsCancelFunc        context.CancelFunc

	ContainerLogs        string
	ContainerLogsLoading bool
	CpuHistory           DataHistory
	MemoryHistory        DataHistory
	NetworkRxHistory     DataHistory
	NetworkTxHistory     DataHistory
	DiskReadHistory      DataHistory
	DiskWriteHistory     DataHistory
	ContainerHistories   map[string]ContainerHistory // History per container ID
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

func (p *ContainerLogsPagination) Clear() {
	p.Lines = []string{}
	p.CurrentPage = 1
	p.TotalPages = 1
}
