package widgets

import (
	"math"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	system "github.com/System-Pulse/server-pulse/system/app"
	model "github.com/System-Pulse/server-pulse/widgets/model"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) setLiveContainerLogs(line string) {
	if m.Monitor.ContainerLogsPagination.Lines == nil {
		m.Monitor.ContainerLogsPagination.Lines = []string{}
		m.Monitor.ContainerLogsPagination.PageSize = 50
	}

	m.Monitor.ContainerLogsPagination.Lines = append(m.Monitor.ContainerLogsPagination.Lines, line)

	maxLines := 1000
	if len(m.Monitor.ContainerLogsPagination.Lines) > maxLines {
		m.Monitor.ContainerLogsPagination.Lines = m.Monitor.ContainerLogsPagination.Lines[len(m.Monitor.ContainerLogsPagination.Lines)-maxLines:]
	}

	total := len(m.Monitor.ContainerLogsPagination.Lines)
	if m.Monitor.ContainerLogsPagination.PageSize <= 0 {
		m.Monitor.ContainerLogsPagination.PageSize = 50
	}

	m.Monitor.ContainerLogsPagination.TotalPages = max(1, int(math.Ceil(float64(total)/float64(m.Monitor.ContainerLogsPagination.PageSize))))

	m.Monitor.ContainerLogsPagination.CurrentPage = m.Monitor.ContainerLogsPagination.TotalPages

	m.updateLogsViewport()
}

func (m Model) readNextLogLine() tea.Cmd {
	return func() tea.Msg {
		if m.Monitor.ContainerLogsChan == nil {
			return system.ContainerLogsStopMsg{ContainerID: ""}
		}

		select {
		case line, ok := <-m.Monitor.ContainerLogsChan:
			if !ok {
				return system.ContainerLogsStopMsg{ContainerID: m.Monitor.SelectedContainer.ID}
			}
			return system.ContainerLogLineMsg{
				ContainerID: m.Monitor.SelectedContainer.ID,
				Line:        line,
			}
		default:
			return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return m.readNextLogLine()()
			})()
		}
	}
}

func (m Model) handleLogLineMsg(msg system.ContainerLogLineMsg) (tea.Model, tea.Cmd) {
	if m.Monitor.SelectedContainer != nil &&
		m.Monitor.SelectedContainer.ID == msg.ContainerID &&
		m.Ui.State == model.StateContainerLogs &&
		m.Monitor.ContainerLogsStreaming {

		m.setLiveContainerLogs(msg.Line)

		return m, m.readNextLogLine()
	}
	return m, nil
}

func (m Model) handleLogsStopMsg() (tea.Model, tea.Cmd) {
	if m.Monitor.LogsCancelFunc != nil {
		m.Monitor.LogsCancelFunc()
	}
	m.Monitor.ContainerLogsStreaming = false
	m.Monitor.ContainerLogsChan = nil
	m.Monitor.LogsCancelFunc = nil

	return m, nil
}

// static logs

func (m *Model) setContainerLogs(logs string) {
	lines := strings.Split(strings.TrimSpace(logs), "\n")

	m.Monitor.ContainerLogsPagination.Lines = lines
	m.Monitor.ContainerLogsPagination.PageSize = 12 // lines by page
	m.Monitor.ContainerLogsPagination.CurrentPage = 1

	total := len(lines)
	if total == 0 {
		m.Monitor.ContainerLogsPagination.TotalPages = 1
	} else {
		m.Monitor.ContainerLogsPagination.TotalPages =
			(int(math.Ceil(float64(total) / float64(m.Monitor.ContainerLogsPagination.PageSize))))
		m.Monitor.ContainerLogsPagination.CurrentPage = m.Monitor.ContainerLogsPagination.TotalPages
	}

	m.updateLogsViewport()
}

func (m *Model) updateLogsViewport() {
	if len(m.Monitor.ContainerLogsPagination.Lines) == 0 {
		m.LogsViewport.SetContent("No logs available")
		return
	}

	if m.Monitor.ContainerLogsPagination.PageSize <= 0 {
		m.Monitor.ContainerLogsPagination.PageSize = 100
	}

	if m.Monitor.ContainerLogsPagination.CurrentPage < 1 {
		m.Monitor.ContainerLogsPagination.CurrentPage = 1
	}
	if m.Monitor.ContainerLogsPagination.CurrentPage > m.Monitor.ContainerLogsPagination.TotalPages {
		m.Monitor.ContainerLogsPagination.CurrentPage = m.Monitor.ContainerLogsPagination.TotalPages
	}

	start := (m.Monitor.ContainerLogsPagination.CurrentPage - 1) * m.Monitor.ContainerLogsPagination.PageSize
	end := min(start+m.Monitor.ContainerLogsPagination.PageSize, len(m.Monitor.ContainerLogsPagination.Lines))

	pageContent := strings.Join(m.Monitor.ContainerLogsPagination.Lines[start:end], "\n")
	m.LogsViewport.SetContent(pageContent)
}

func (m Model) handleLogsStreamMsg(msg system.ContainerLogsStreamMsg) (tea.Model, tea.Cmd) {
	m.Monitor.ContainerLogsStreaming = true
	m.Monitor.ContainerLogsLoading = false
	m.Monitor.ContainerLogsChan = msg.LogChan
	m.Monitor.LogsCancelFunc = msg.CancelFunc

	m.Monitor.ContainerLogsPagination.Clear()
	m.updateLogsViewport()

	return m, m.readNextLogLine()
}

func (m *Model) handleRealTimeStats(containerID string, statsChan chan app.ContainerStatsMsg) {
	for stats := range statsChan {
		if m.Monitor.ContainerDetails != nil && m.Monitor.ContainerDetails.ID == containerID {
			m.Monitor.ContainerDetails.Stats.CPUPercent = stats.CPUPercent
			m.Monitor.ContainerDetails.Stats.MemoryPercent = stats.MemPercent
			m.Monitor.ContainerDetails.Stats.MemoryUsage = stats.MemUsage
			m.Monitor.ContainerDetails.Stats.MemoryLimit = stats.MemLimit
			m.Monitor.ContainerDetails.Stats.NetworkRx = stats.NetRX
			m.Monitor.ContainerDetails.Stats.NetworkTx = stats.NetTX

			m.updateChartsWithStats(stats)
		}
	}
}
