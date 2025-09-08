package widgets

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.err != nil {
		return fmt.Sprintf("Erreur: %v\n", m.err)
	}

	// Checking the minimum terminal size
	if m.width < 40 || m.height < 10 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render("Terminal too small!\nMinimum size: 40x10\nCurrent: " +
				fmt.Sprintf("%dx%d", m.width, m.height))
	}

	var currentView string

	if m.activeView != -1 {
		switch m.activeView {
		case 0: // Monitor
			if m.isMonitorActive {
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

	hasSpecialView := m.confirmationVisible ||
		m.containerViewState == ContainerViewSingle ||
		m.containerViewState == ContainerViewLogs ||
		(m.isMonitorActive && m.selectedMonitor == 2 && m.containerMenuState == ContainerMenuVisible)

	if hasSpecialView {
		if m.confirmationVisible {
			baseContent := ""
			if m.containerViewState == ContainerViewSingle {
				baseContent = m.renderContainerSingleView()
			} else if m.containerViewState == ContainerViewLogs {
				baseContent = m.renderContainerLogs()
			} else if m.isMonitorActive && m.selectedMonitor == 2 {
				baseContent = m.renderContainersTable()
			} else {
				baseContent = currentView
			}
			confirmationDialog := m.renderConfirmationDialog()
			mainContent = lipgloss.JoinVertical(lipgloss.Center, baseContent, confirmationDialog)
		} else if m.containerViewState == ContainerViewSingle {
			mainContent = m.renderContainerSingleView()
		} else if m.containerViewState == ContainerViewLogs {
			mainContent = m.renderContainerLogs()
		} else if m.isMonitorActive && m.selectedMonitor == 2 {
			if m.containerMenuState == ContainerMenuVisible {
				mainContent = m.renderContainerMenu()
			} else {
				mainContent = m.renderContainersTable()
			}
		}
	} else {
		if m.isMonitorActive && m.selectedMonitor == 1 {
			if m.searchMode {
				searchBar := searchBarStyle.
					Render(m.searchInput.View())
				currentView = lipgloss.JoinVertical(lipgloss.Left, searchBar, m.processTable.View())
			} else {
				currentView = m.processTable.View()
			}
		}

		m.viewport.SetContent(currentView)
		mainContent = m.viewport.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		home,
		tabs,
		mainContent,
		footer,
	)
}
