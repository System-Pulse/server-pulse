package widgets

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"

	model "github.com/System-Pulse/server-pulse/widgets/model"

	"github.com/System-Pulse/server-pulse/system/app"
	resource "github.com/System-Pulse/server-pulse/system/resource"
)


const (
	ContainerMenuHidden model.ContainerMenuState = iota
	ContainerMenuVisible
)

const (
	ContainerViewNone model.ContainerViewState = iota
	ContainerViewSingle
	ContainerViewLogs
	ContainerViewConfirmation
)

type Model struct {
	Network             resource.NetworkInfo
	Err                 error
	Monitor             model.MonitorModel
	Ui                  model.UIModel
	ProgressBars        map[string]progress.Model
	ContainerTab        model.ContainerTab
	LastUpdate          time.Time
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
