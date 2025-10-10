package model

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type AppState string

const (
	StateHome               AppState = "home"
	StateMonitor            AppState = "monitor"
	StateSystem             AppState = "monitor.system"
	StateProcess            AppState = "monitor.process"
	StateContainers         AppState = "monitor.containers"
	StateContainer          AppState = "monitor.containers.single"
	StateContainerLogs      AppState = "monitor.containers.logs"
	StateDiagnostics        AppState = "diagnostics"
	StateCertificateDetails AppState = "diagnostics.certificate"
	StateSSHRootDetails     AppState = "diagnostics.sshroot"
	StateOpenedPortsDetails AppState = "diagnostics.openedports"
	StateFirewallDetails    AppState = "diagnostics.firewall"
	StateAutoBanDetails     AppState = "diagnostics.autoban"
	StateLogDetails         AppState = "diagnostics.logs"
	StateLogEntryDetails    AppState = "diagnostics.logs.entry"
	StateNetwork            AppState = "network"
	StateInterfaces         AppState = "network.interfaces"
	StateConnectivity       AppState = "network.connectivity"
	StateConfiguration      AppState = "network.configuration"
	StateProtocolAnalysis   AppState = "network.protocol.analysis"
	StateReporting          AppState = "reporting"
	StatePerformance        AppState = "performance"
	StateSystemHealth       AppState = "performance.systemhealth"
	StateInputOutput        AppState = "performance.inputoutput"
	StateCPU                AppState = "performance.cpu"
	StateMemory             AppState = "performance.memory"
	StateQuickTests         AppState = "performance.quicktests"
)

type UIModel struct {
	State           AppState
	PreviousState   AppState
	Loading         bool
	SelectedTab     int
	SelectedMonitor int
	IsMonitorActive bool
	IsNetworkActive bool
	ActiveView      int
	SearchInput     textinput.Model
	SearchMode      bool
	Tabs            Menu
	Width           int
	Height          int
	Ready           bool
	Spinner         spinner.Model
	Viewport        viewport.Model
	ContentHeight   int
}
