package widgets

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.err != nil {
		return fmt.Sprintf("Erreur: %v\n", m.err)
	}

	// Si la vue unique du conteneur est active, l'afficher en plein écran
	if m.containerSingleView.Visible {
		return m.renderContainerSingleView()
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
		currentView = m.renderSystem()
	}

	home := m.renderHome()
	tabs := m.renderTabs()
	footer := m.renderFooter()

	var mainContent string
	if m.isMonitorActive && m.selectedMonitor == 1 {
		if m.searchMode {
			searchBar := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("57")).
				Padding(0, 1).
				MarginBottom(1).
				Render(m.searchInput.View())
			mainContent = lipgloss.JoinVertical(lipgloss.Left, searchBar, m.processTable.View())
		} else {
			mainContent = m.processTable.View()
		}
	} else {
		if m.activeView != -1 {
			m.viewport.SetContent(currentView)
		}
		mainContent = m.viewport.View()
	}

	mainView := lipgloss.JoinVertical(lipgloss.Left,
		home,
		tabs,
		mainContent,
		footer,
	)

	// Superposer le menu si visible
	if m.containerMenu.Visible {
		menu := m.renderContainerMenu()
		// Utiliser Place pour superposer le menu au centre
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Left, mainView, menu),
			lipgloss.WithWhitespaceChars(""),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")))
	}

	return mainView
}
