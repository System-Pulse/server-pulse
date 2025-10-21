package widgets

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"

	model "github.com/System-Pulse/server-pulse/widgets/model"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/system/security"
)

type Model struct {
	Network             model.NetworkModel
	Err                 error
	Monitor             model.MonitorModel
	Ui                  model.UIModel
	ProgressBars        map[string]progress.Model
	ContainerTab        model.ContainerTab
	LogsViewport        viewport.Model
	EnableAnimations    bool
	LastChartUpdate     time.Time
	LastOperationMsg    string
	OperationInProgress bool
	ConfirmationVisible bool
	ConfirmationMessage string
	ConfirmationAction  string
	ConfirmationData    any
	ScrollSensitivity   int
	LastScrollTime      time.Time
	MouseEnabled        bool
	Diagnostic          model.DiagnosticModel
	SecurityManager     *security.SecurityManager
	HelpSystem          HelpSystem
	AsRoot              bool
	SudoAvailable       bool
	CanRunSudo          bool
}

type connectionStats struct {
	tcp         int
	udp         int
	listening   int
	established int
}

func (m *Model) setState(newState model.AppState) {
	if m.Ui.State != newState {
		m.Ui.PreviousState = m.Ui.State
		m.Ui.State = newState
	}

	switch newState {
	case model.StateHome:
		m.Ui.SelectedTab = m.Ui.ActiveView
		m.Ui.ActiveView = -1
	case model.StateMonitor, model.StateSystem, model.StateProcess, model.StateContainers,
		model.StateContainer, model.StateContainerLogs:
		m.Ui.SelectedTab = 0
	case model.StateDiagnostics, model.StateCertificateDetails:
		m.Ui.SelectedTab = 1
	case model.StateNetwork:
		m.Ui.SelectedTab = 2
	case model.StateReporting:
		m.Ui.SelectedTab = 3
	}

	switch newState {
	case model.StateSystem:
		m.Ui.SelectedMonitor = 0
	case model.StateProcess:
		m.Ui.SelectedMonitor = 1
	case model.StateContainers, model.StateContainer, model.StateContainerLogs:
		m.Ui.SelectedMonitor = 2
	}
}

func (m *Model) goBack() {
	m.cleanupLogsStream()
	switch m.Ui.State {
	case model.StateContainer, model.StateContainerLogs:
		m.stopContainerStats()
		m.setState(model.StateContainers)
	case model.StateMonitor, model.StateDiagnostics, model.StateNetwork, model.StateReporting,
		model.StateSystem, model.StateProcess, model.StateContainers:
		m.stopContainerStats()
		m.setState(model.StateHome)
	default:
		if m.Ui.PreviousState != "" {
			m.stopContainerStats()
			m.setState(m.Ui.PreviousState)
		} else {
			m.stopContainerStats()
			m.setState(model.StateHome)
		}
	}
}

// Methods to support main application loop
func (m Model) GetPendingShellExec() *model.ShellExecRequest {
	return m.Monitor.PendingShellExec
}

func (m Model) ShouldQuit() bool {
	return m.Monitor.ShouldQuit
}

func (m *Model) ClearPendingShellExec() {
	m.Monitor.PendingShellExec = nil
}

func (m Model) GetApp() *app.DockerManager {
	return m.Monitor.App
}
