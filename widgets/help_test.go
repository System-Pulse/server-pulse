package widgets

import (
	"testing"

	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/bubbles/key"
)

func TestHelpSystemCreation(t *testing.T) {
	helpSystem := NewHelpSystem()

	if helpSystem.HelpModel.Width != 0 {
		t.Errorf("Expected initial help model width to be 0, got %d", helpSystem.HelpModel.Width)
	}

	if helpSystem.ShowAll != false {
		t.Errorf("Expected initial ShowAll to be false, got %v", helpSystem.ShowAll)
	}
}

func TestHelpSystemToggle(t *testing.T) {
	helpSystem := NewHelpSystem()

	// Test initial state
	if helpSystem.ShowAll != false {
		t.Errorf("Expected initial ShowAll to be false, got %v", helpSystem.ShowAll)
	}

	// Test toggle
	helpSystem.ToggleHelp()
	if helpSystem.ShowAll != true {
		t.Errorf("Expected ShowAll to be true after toggle, got %v", helpSystem.ShowAll)
	}

	// Test toggle back
	helpSystem.ToggleHelp()
	if helpSystem.ShowAll != false {
		t.Errorf("Expected ShowAll to be false after second toggle, got %v", helpSystem.ShowAll)
	}
}

func TestHelpSystemSetWidth(t *testing.T) {
	helpSystem := NewHelpSystem()

	helpSystem.SetWidth(100)
	if helpSystem.HelpModel.Width != 100 {
		t.Errorf("Expected help model width to be 100, got %d", helpSystem.HelpModel.Width)
	}
}

func TestGetKeyMapForState(t *testing.T) {
	helpSystem := NewHelpSystem()

	// Test home state
	homeKeyMap := helpSystem.GetKeyMapForState(model.StateHome, model.ContainerTab(0))
	if homeKeyMap == nil {
		t.Error("Expected non-nil keymap for home state")
	}

	// Test system state
	systemKeyMap := helpSystem.GetKeyMapForState(model.StateSystem, model.ContainerTab(0))
	if systemKeyMap == nil {
		t.Error("Expected non-nil keymap for system state")
	}

	// Test process state
	processKeyMap := helpSystem.GetKeyMapForState(model.StateProcess, model.ContainerTab(0))
	if processKeyMap == nil {
		t.Error("Expected non-nil keymap for process state")
	}

	// Test unknown state should return base keymap
	unknownKeyMap := helpSystem.GetKeyMapForState(model.AppState("unknown"), model.ContainerTab(0))
	if unknownKeyMap == nil {
		t.Error("Expected non-nil keymap for unknown state")
	}
}

func TestKeyMapInterfaces(t *testing.T) {
	helpSystem := NewHelpSystem()

	// Test that all keymaps implement the interface properly
	states := []model.AppState{
		model.StateHome,
		model.StateMonitor,
		model.StateSystem,
		model.StateProcess,
		model.StateContainers,
		model.StateContainer,
		model.StateContainerLogs,
		model.StateNetwork,
		model.StateDiagnostics,
		model.StateReporting,
	}

	for _, state := range states {
		keyMap := helpSystem.GetKeyMapForState(state, model.ContainerTab(0))

		// Test ShortHelp returns []key.Binding
		shortHelp := keyMap.ShortHelp()
		if shortHelp == nil {
			t.Errorf("Expected non-nil ShortHelp for state %v", state)
		}

		// Test FullHelp returns [][]key.Binding
		fullHelp := keyMap.FullHelp()
		if fullHelp == nil {
			t.Errorf("Expected non-nil FullHelp for state %v", state)
		}

		// Verify that the bindings are valid
		for _, binding := range shortHelp {
			if binding.Help().Key == "" {
				t.Errorf("Empty key in ShortHelp binding for state %v", state)
			}
		}

		for _, column := range fullHelp {
			for _, binding := range column {
				if binding.Help().Key == "" {
					t.Errorf("Empty key in FullHelp binding for state %v", state)
				}
			}
		}
	}
}

func TestHelpView(t *testing.T) {
	helpSystem := NewHelpSystem()

	// Test short help view
	shortView := helpSystem.View(model.StateHome, model.ContainerTab(0))
	if shortView == "" {
		t.Error("Expected non-empty short help view")
	}

	// Test full help view
	helpSystem.ToggleHelp()
	fullView := helpSystem.View(model.StateHome, model.ContainerTab(0))
	if fullView == "" {
		t.Error("Expected non-empty full help view")
	}

	// Views should be different
	if shortView == fullView {
		t.Error("Expected short and full help views to be different")
	}
}

func TestDiagnosticHelpVariations(t *testing.T) {
	helpSystem := NewHelpSystem()

	// Test different diagnostic tabs have different help
	securityHelp := helpSystem.View(model.StateDiagnostics, model.DiagnosticSecurityChecks)
	logsHelp := helpSystem.View(model.StateDiagnostics, model.DiagnosticTabLogs)
	performanceHelp := helpSystem.View(model.StateDiagnostics, model.DiagnosticTabPerformances)

	if securityHelp == "" {
		t.Error("Expected non-empty help for security tab")
	}
	if logsHelp == "" {
		t.Error("Expected non-empty help for logs tab")
	}
	if performanceHelp == "" {
		t.Error("Expected non-empty help for performance tab")
	}

	// Help should be different for different tabs
	if securityHelp == logsHelp {
		t.Error("Expected different help for security and logs tabs")
	}
	if securityHelp == performanceHelp {
		t.Error("Expected different help for security and performance tabs")
	}
}

func TestBaseKeyMap(t *testing.T) {
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

	// Test ShortHelp
	shortHelp := baseKeys.ShortHelp()
	if len(shortHelp) != 2 {
		t.Errorf("Expected 2 bindings in ShortHelp, got %d", len(shortHelp))
	}

	// Test FullHelp
	fullHelp := baseKeys.FullHelp()
	if len(fullHelp) != 1 {
		t.Errorf("Expected 1 column in FullHelp, got %d", len(fullHelp))
	}
	if len(fullHelp[0]) != 3 {
		t.Errorf("Expected 3 bindings in FullHelp column, got %d", len(fullHelp[0]))
	}
}
