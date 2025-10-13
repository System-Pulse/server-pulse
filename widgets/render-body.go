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
	case model.StateCertificateDetails:
		currentView = m.renderCertificateDetails()
	case model.StateSSHRootDetails:
		currentView = m.renderSSHRootDetails()
	case model.StateOpenedPortsDetails:
		currentView = m.renderOpenedPortsDetails()
	case model.StateFirewallDetails:
		currentView = m.renderFirewallDetails()
	case model.StateAutoBanDetails:
		currentView = m.renderAutoBanDetails()
	case model.StateReporting:
		currentView = m.renderReporting()
	default:
		currentView = fmt.Sprintf("Unknown state: %v", m.Ui.State)
	}

	return currentView
}
