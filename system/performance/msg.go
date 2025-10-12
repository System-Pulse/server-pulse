package performance

import tea "github.com/charmbracelet/bubbletea"

type GetHealthMetricsMsg struct{}

type HealthMetricsMsg struct {
	Metrics *HealthMetrics
	Score   *HealthScore
}

func GetHealthMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		return GetHealthMetricsMsg{}
	}
}
