package widgets

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.Ui.Ready {
		return "Initializing..."
	}
	if m.Err != nil {
		return fmt.Sprintf("Erreur: %v\n", m.Err)
	}

	// Checking the minimum terminal size
	if m.Ui.Width < 40 || m.Ui.Height < 10 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render("Terminal too small!\nMinimum size: 40x10\nCurrent: " +
				fmt.Sprintf("%dx%d", m.Ui.Width, m.Ui.Height))
	}

	var currentView string

	if m.Ui.ActiveView != -1 {
		switch m.Ui.ActiveView {
		case 0: // Monitor
			if m.Ui.IsMonitorActive {
				currentView = m.renderMonitor()
			} else {
				currentView = m.renderSystem()
			}
		case 1: // Diagnostic
			currentView = m.renderDignostics()
		case 2: // Network
			currentView = m.renderNetwork()
		case 3: // Reporting
			currentView = m.renderReporting()
		}
	} else {
		// currentView = m.renderSystem()
	}

	home := m.renderHome()
	tabs := m.renderTabs()
	footer := m.renderFooter()

	var mainContent string

	hasSpecialView := m.ConfirmationVisible ||
		m.Monitor.ContainerViewState == ContainerViewSingle ||
		m.Monitor.ContainerViewState == ContainerViewLogs ||
		(m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 2 && m.Monitor.ContainerMenuState == ContainerMenuVisible)

	if hasSpecialView {
		if m.ConfirmationVisible {
			baseContent := ""
			if m.Monitor.ContainerViewState == ContainerViewSingle {
				baseContent = m.renderContainerSingleView()
			} else if m.Monitor.ContainerViewState == ContainerViewLogs {
				baseContent = m.renderContainerLogs()
			} else if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 2 {
				p := "Search a container..."
				baseContent = m.renderTable(m.Monitor.Container, p)
			} else {
				baseContent = currentView
			}
			confirmationDialog := m.renderConfirmationDialog()
			mainContent = lipgloss.JoinVertical(lipgloss.Center, baseContent, confirmationDialog)
		} else if m.Monitor.ContainerViewState == ContainerViewSingle {
			mainContent = m.renderContainerSingleView()
		} else if m.Monitor.ContainerViewState == ContainerViewLogs {
			mainContent = m.renderContainerLogs()
		} else if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 2 {
			if m.Monitor.ContainerMenuState == ContainerMenuVisible {
				mainContent = m.renderContainerMenu()
			} else {
				p := "Search a process..."
				mainContent = m.renderTable(m.Monitor.ProcessTable,p)
			}
		}
	} else {
		if m.Ui.IsMonitorActive && m.Ui.SelectedMonitor == 1 {
			if m.Ui.SearchMode {
				searchBar := searchBarStyle.
					Render(m.Ui.SearchInput.View())
				currentView = lipgloss.JoinVertical(lipgloss.Left, searchBar, m.Monitor.ProcessTable.View())
			} else {
				currentView = m.Monitor.ProcessTable.View()
			}
		}

		m.Ui.Viewport.SetContent(currentView)
		mainContent = m.Ui.Viewport.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		home,
		tabs,
		mainContent,
		footer,
	)
}
