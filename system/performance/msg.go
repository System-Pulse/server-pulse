package performance

import (
	"github.com/System-Pulse/server-pulse/widgets/model"
	tea "github.com/charmbracelet/bubbletea"
)

type GetHealthMetricsMsg struct{}

type HealthMetricsMsg struct {
	Metrics *HealthMetrics
	Score   *HealthScore
}

type GetIOMetricsMsg struct{}

type IOMetricsMsg struct {
	Metrics *model.IOMetrics
	Error   error
}

type GetCPUMetricsMsg struct{}

type CPUMetricsMsg struct {
	Metrics *model.CPUMetrics
	Error   error
}

type GetMemoryMetricsMsg struct{}

type MemoryMetricsMsg struct {
	Metrics *model.MemoryMetrics
	Error   error
}

func GetHealthMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		return GetHealthMetricsMsg{}
	}
}

func GetIOMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		return GetIOMetricsMsg{}
	}
}

func GetCPUMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		return GetCPUMetricsMsg{}
	}
}

func GetMemoryMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		return GetMemoryMetricsMsg{}
	}
}
