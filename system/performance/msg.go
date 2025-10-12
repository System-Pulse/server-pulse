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
