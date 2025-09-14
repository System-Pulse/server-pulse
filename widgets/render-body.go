package widgets

import (
	"fmt"

	"github.com/System-Pulse/server-pulse/widgets/model"
)

func (m Model) renderMainContent() string {
	var currentView string

	switch m.Ui.State {
	case model.StateHome:
		currentView = m.renderSystem()
	case model.StateMonitor:
		currentView = m.renderMonitor()
	case model.StateSystem:
		currentView = m.renderSystem()
	case model.StateProcess:
		currentView = m.renderProcesses()
	case model.StateContainers:
		currentView = m.renderContainers()
	case model.StateContainer:
		currentView = m.renderContainerSingleView()
	case model.StateContainerLogs:
		currentView = m.renderContainerLogs()
	case model.StateNetwork:
		currentView = m.renderNetwork()
	case model.StateDiagnostics:
		currentView = m.renderDignostics()
	case model.StateReporting:
		currentView = m.renderReporting()
	default:
		currentView = fmt.Sprintf("Unknown state: %v", m.Ui.State)
	}

	// Utilise le viewport pour le contenu scrollable
	// switch m.Ui.State {
	// case model.StateSystem, model.StateContainerLogs, model.StateHome, model.StateMonitor:
	m.Ui.Viewport.SetContent(currentView)
	return m.Ui.Viewport.View()
	// }

	// return currentView
}
