package widgets

import (
	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap defines the interface for state-specific keybindings
type KeyMap interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

// BaseKeyMap contains keys that are available in most states
type BaseKeyMap struct {
	Help key.Binding
	Quit key.Binding
	Back key.Binding
}

func (k BaseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k BaseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit, k.Back},
	}
}

// HomeKeyMap defines keybindings for the home state
type HomeKeyMap struct {
	BaseKeyMap
	Select   key.Binding
	Navigate key.Binding
	Quick1   key.Binding
	Quick2   key.Binding
	Quick3   key.Binding
	Quick4   key.Binding
}

func (k HomeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Navigate, k.Help, k.Quit}
}

func (k HomeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Navigate},
		{k.Quick1, k.Quick2, k.Quick3, k.Quick4},
		{k.Help, k.Quit},
	}
}

// MonitorKeyMap defines keybindings for monitor states
type MonitorKeyMap struct {
	BaseKeyMap
	Select   key.Binding
	Navigate key.Binding
	Quick1   key.Binding
	Quick2   key.Binding
	Quick3   key.Binding
}

func (k MonitorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Navigate, k.Help, k.Back, k.Quit}
}

func (k MonitorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Navigate},
		{k.Quick1, k.Quick2, k.Quick3},
		{k.Help, k.Back, k.Quit},
	}
}

// SystemKeyMap defines keybindings for system state
type SystemKeyMap struct {
	BaseKeyMap
	Scroll key.Binding
}

func (k SystemKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Scroll, k.Help, k.Back, k.Quit}
}

func (k SystemKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Scroll},
		{k.Help, k.Back, k.Quit},
	}
}

// ProcessKeyMap defines keybindings for process state
type ProcessKeyMap struct {
	BaseKeyMap
	Navigate key.Binding
	Search   key.Binding
	Kill     key.Binding
	SortMem  key.Binding
	SortCPU  key.Binding
}

func (k ProcessKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Navigate, k.Search, k.Help, k.Back, k.Quit}
}

func (k ProcessKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigate, k.Search},
		{k.Kill, k.SortMem, k.SortCPU},
		{k.Help, k.Back, k.Quit},
	}
}

// ContainersKeyMap defines keybindings for containers state
type ContainersKeyMap struct {
	BaseKeyMap
	Navigate key.Binding
	Select   key.Binding
	Search   key.Binding
}

func (k ContainersKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Navigate, k.Select, k.Search, k.Help, k.Back, k.Quit}
}

func (k ContainersKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigate, k.Select, k.Search},
		{k.Help, k.Back, k.Quit},
	}
}

// ContainerKeyMap defines keybindings for single container state
type ContainerKeyMap struct {
	BaseKeyMap
	SwitchTab key.Binding
}

func (k ContainerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SwitchTab, k.Help, k.Back, k.Quit}
}

func (k ContainerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.SwitchTab},
		{k.Help, k.Back, k.Quit},
	}
}

// ContainerLogsKeyMap defines keybindings for container logs state
type ContainerLogsKeyMap struct {
	BaseKeyMap
	Scroll  key.Binding
	Refresh key.Binding
	HomeEnd key.Binding
	Stream  key.Binding
}

func (k ContainerLogsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Scroll, k.Refresh, k.Help, k.Back, k.Quit}
}

func (k ContainerLogsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Scroll, k.Refresh, k.HomeEnd},
		{k.Stream},
		{k.Help, k.Back, k.Quit},
	}
}

// NetworkKeyMap defines keybindings for network state
type NetworkKeyMap struct {
	BaseKeyMap
	SwitchTab key.Binding
	Ping      key.Binding
	Trace     key.Binding
	Clear     key.Binding
	Switch    key.Binding
	Navigate  key.Binding
}

func (k NetworkKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SwitchTab, k.Navigate, k.Help, k.Back, k.Quit}
}

func (k NetworkKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.SwitchTab, k.Navigate},
		{k.Ping, k.Trace, k.Clear, k.Switch},
		{k.Help, k.Back, k.Quit},
	}
}

// DiagnosticsKeyMap defines keybindings for diagnostics state
type DiagnosticsKeyMap struct {
	BaseKeyMap
	TimeRange    key.Binding
	Level        key.Binding
	SwitchFilter key.Binding
	SwitchTab    key.Binding
	Reload       key.Binding
	Search       key.Binding
	Service      key.Binding
	Navigate     key.Binding
	Details      key.Binding
	Security     key.Binding
	Performance  key.Binding
	Logs         key.Binding
}

func (k DiagnosticsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Navigate, k.Search, k.SwitchTab, k.Help, k.Back, k.Quit}
}

func (k DiagnosticsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigate, k.Search, k.Service, k.SwitchTab},
		{k.TimeRange, k.Level, k.SwitchFilter, k.Reload},
		{k.Security, k.Performance, k.Logs, k.Details},
		{k.Help, k.Back, k.Quit},
	}
}

// ConfirmationKeyMap defines keybindings for confirmation dialogs
type ConfirmationKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func (k ConfirmationKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel}
}

func (k ConfirmationKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Cancel},
	}
}

// ContainerMenuKeyMap defines keybindings for container menu
type ContainerMenuKeyMap struct {
	BaseKeyMap
	Navigate key.Binding
	Select   key.Binding
	Open     key.Binding
	Logs     key.Binding
	Restart  key.Binding
	Delete   key.Binding
	Remove   key.Binding
	Toggle   key.Binding
	Pause    key.Binding
	Exec     key.Binding
}

func (k ContainerMenuKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Navigate, k.Select, k.Help, k.Back, k.Quit}
}

func (k ContainerMenuKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigate, k.Select},
		{k.Open, k.Logs, k.Restart, k.Delete},
		{k.Remove, k.Toggle, k.Pause, k.Exec},
		{k.Help, k.Back, k.Quit},
	}
}

// HelpSystem manages the help system state
type HelpSystem struct {
	HelpModel help.Model
	ShowAll   bool
}

// NewHelpSystem creates a new help system
func NewHelpSystem() HelpSystem {
	helpModel := help.New()
	helpModel.Styles = help.Styles{
		ShortKey:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		ShortDesc:      lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		ShortSeparator: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Ellipsis:       lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		FullKey:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		FullDesc:       lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		FullSeparator:  lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
	}

	return HelpSystem{
		HelpModel: helpModel,
		ShowAll:   false,
	}
}

// GetKeyMapForState returns the appropriate keymap for the current state
func (hs *HelpSystem) GetKeyMapForState(state model.AppState, diagnosticSelectedItem model.ContainerTab) KeyMap {
	baseKeys := BaseKeyMap{
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "back"),
		),
	}

	switch state {
	case model.StateHome:
		return HomeKeyMap{
			BaseKeyMap: baseKeys,
			Select: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			Navigate: key.NewBinding(
				key.WithKeys("tab", "left", "right"),
				key.WithHelp("tab/←→", "navigate"),
			),
			Quick1: key.NewBinding(
				key.WithKeys("1"),
				key.WithHelp("1", "monitor"),
			),
			Quick2: key.NewBinding(
				key.WithKeys("2"),
				key.WithHelp("2", "diagnostics"),
			),
			Quick3: key.NewBinding(
				key.WithKeys("3"),
				key.WithHelp("3", "network"),
			),
			Quick4: key.NewBinding(
				key.WithKeys("4"),
				key.WithHelp("4", "reporting"),
			),
		}

	case model.StateMonitor:
		return MonitorKeyMap{
			BaseKeyMap: baseKeys,
			Select: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			Navigate: key.NewBinding(
				key.WithKeys("tab", "left", "right"),
				key.WithHelp("tab/←→", "navigate"),
			),
			Quick1: key.NewBinding(
				key.WithKeys("1"),
				key.WithHelp("1", "system"),
			),
			Quick2: key.NewBinding(
				key.WithKeys("2"),
				key.WithHelp("2", "process"),
			),
			Quick3: key.NewBinding(
				key.WithKeys("3"),
				key.WithHelp("3", "containers"),
			),
		}

	case model.StateSystem:
		return SystemKeyMap{
			BaseKeyMap: baseKeys,
			Scroll: key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp("↑↓", "scroll"),
			),
		}

	case model.StateProcess:
		return ProcessKeyMap{
			BaseKeyMap: baseKeys,
			Navigate: key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp("↑↓", "navigate"),
			),
			Search: key.NewBinding(
				key.WithKeys("/"),
				key.WithHelp("/", "search"),
			),
			Kill: key.NewBinding(
				key.WithKeys("k"),
				key.WithHelp("k", "kill"),
			),
			SortMem: key.NewBinding(
				key.WithKeys("m"),
				key.WithHelp("m", "sort by mem"),
			),
			SortCPU: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "sort by cpu"),
			),
		}

	case model.StateContainers:
		return ContainersKeyMap{
			BaseKeyMap: baseKeys,
			Navigate: key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp("↑↓", "navigate"),
			),
			Select: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "menu"),
			),
			Search: key.NewBinding(
				key.WithKeys("/"),
				key.WithHelp("/", "search"),
			),
		}

	case model.StateContainer:
		return ContainerKeyMap{
			BaseKeyMap: baseKeys,
			SwitchTab: key.NewBinding(
				key.WithKeys("tab", "left", "right"),
				key.WithHelp("tab/←→", "switch tabs"),
			),
		}

	case model.StateContainerLogs:
		return ContainerLogsKeyMap{
			BaseKeyMap: baseKeys,
			Scroll: key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp("↑↓", "scroll"),
			),
			Refresh: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "refresh"),
			),
			HomeEnd: key.NewBinding(
				key.WithKeys("home", "end"),
				key.WithHelp("home/end", "navigate"),
			),
			Stream: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "toggle stream"),
			),
		}

	case model.StateNetwork:
		return NetworkKeyMap{
			BaseKeyMap: baseKeys,
			SwitchTab: key.NewBinding(
				key.WithKeys("tab", "left", "right"),
				key.WithHelp("tab/←→", "switch tabs"),
			),
			Ping: key.NewBinding(
				key.WithKeys("p"),
				key.WithHelp("p", "ping"),
			),
			Trace: key.NewBinding(
				key.WithKeys("t"),
				key.WithHelp("t", "trace route"),
			),
			Clear: key.NewBinding(
				key.WithKeys("c"),
				key.WithHelp("c", "clear"),
			),
			Switch: key.NewBinding(
				key.WithKeys("space"),
				key.WithHelp("space", "switch view"),
			),
			Navigate: key.NewBinding(
				key.WithKeys("up", "down"),
				key.WithHelp("↑↓", "navigate"),
			),
		}

	case model.StateDiagnostics, model.StateCertificateDetails, model.StateSSHRootDetails:
		switch diagnosticSelectedItem {
		case model.DiagnosticSecurityChecks:
			return DiagnosticsKeyMap{
				BaseKeyMap: baseKeys,
				SwitchTab: key.NewBinding(
					key.WithKeys("left", "right"),
					key.WithHelp("←→", "switch tabs"),
				),
				Reload: key.NewBinding(
					key.WithKeys("enter", "r"),
					key.WithHelp("enter/r", "reload"),
				),
				Search: key.NewBinding(
					key.WithKeys("/"),
					key.WithHelp("/", "search"),
				),
				Navigate: key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑↓", "navigate"),
				),
				Details: key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "details"),
				),
				Security: key.NewBinding(
					key.WithKeys("1"),
					key.WithHelp("1", "security"),
				),
				Performance: key.NewBinding(
					key.WithKeys("2"),
					key.WithHelp("2", "performance"),
				),
				Logs: key.NewBinding(
					key.WithKeys("3"),
					key.WithHelp("3", "logs"),
				),
			}
		case model.DiagnosticTabPerformances:
			return DiagnosticsKeyMap{
				BaseKeyMap: baseKeys,
				SwitchTab: key.NewBinding(
					key.WithKeys("left", "right"),
					key.WithHelp("←→", "switch tabs"),
				),
				Navigate: key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑↓", "navigate"),
				),
				Details: key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "details"),
				),
				Security: key.NewBinding(
					key.WithKeys("1"),
					key.WithHelp("1", "security"),
				),
				Performance: key.NewBinding(
					key.WithKeys("2"),
					key.WithHelp("2", "performance"),
				),
				Logs: key.NewBinding(
					key.WithKeys("3"),
					key.WithHelp("3", "logs"),
				),
			}
		case model.DiagnosticTabLogs:
			return DiagnosticsKeyMap{
				BaseKeyMap: baseKeys,
				TimeRange: key.NewBinding(
					key.WithKeys("shift+left", "shift+right"),
					key.WithHelp("shift+←→", "time range"),
				),
				Level: key.NewBinding(
					key.WithKeys("shift+left", "shift+right"),
					key.WithHelp("shift+←→", "level"),
				),
				SwitchFilter: key.NewBinding(
					key.WithKeys("space"),
					key.WithHelp("space", "switch filter"),
				),
				SwitchTab: key.NewBinding(
					key.WithKeys("left", "right"),
					key.WithHelp("←→", "switch tabs"),
				),
				Reload: key.NewBinding(
					key.WithKeys("enter", "r"),
					key.WithHelp("enter/r", "reload"),
				),
				Search: key.NewBinding(
					key.WithKeys("/"),
					key.WithHelp("/", "search"),
				),
				Service: key.NewBinding(
					key.WithKeys("s"),
					key.WithHelp("s", "service"),
				),
				Navigate: key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑↓", "navigate"),
				),
				Details: key.NewBinding(
					key.WithKeys("d"),
					key.WithHelp("d", "details"),
				),
				Security: key.NewBinding(
					key.WithKeys("1"),
					key.WithHelp("1", "security"),
				),
				Performance: key.NewBinding(
					key.WithKeys("2"),
					key.WithHelp("2", "performance"),
				),
				Logs: key.NewBinding(
					key.WithKeys("3"),
					key.WithHelp("3", "logs"),
				),
			}
		default:
			return DiagnosticsKeyMap{
				BaseKeyMap: baseKeys,
				SwitchTab: key.NewBinding(
					key.WithKeys("left", "right"),
					key.WithHelp("←→", "switch tabs"),
				),
				Navigate: key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑↓", "navigate"),
				),
				Details: key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "details"),
				),
				Security: key.NewBinding(
					key.WithKeys("1"),
					key.WithHelp("1", "security"),
				),
				Performance: key.NewBinding(
					key.WithKeys("2"),
					key.WithHelp("2", "performance"),
				),
				Logs: key.NewBinding(
					key.WithKeys("3"),
					key.WithHelp("3", "logs"),
				),
			}
		}

	case model.StateReporting:
		return BaseKeyMap{
			Help: baseKeys.Help,
			Quit: baseKeys.Quit,
			Back: baseKeys.Back,
		}
	}

	return baseKeys
}

// ToggleHelp toggles between short and full help view
func (hs *HelpSystem) ToggleHelp() {
	hs.ShowAll = !hs.ShowAll
}

// View returns the help view for the current state
func (hs *HelpSystem) View(state model.AppState, diagnosticSelectedItem model.ContainerTab) string {
	keyMap := hs.GetKeyMapForState(state, diagnosticSelectedItem)
	if hs.ShowAll {
		return hs.HelpModel.FullHelpView(keyMap.FullHelp())
	}
	return hs.HelpModel.ShortHelpView(keyMap.ShortHelp())
}

// SetWidth sets the width for the help model
func (hs *HelpSystem) SetWidth(width int) {
	hs.HelpModel.Width = width
}
