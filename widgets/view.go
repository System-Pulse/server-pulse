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

	// Vérifier si on est en vue détaillée d'un conteneur
	if m.containerViewState == ContainerViewSingle {
		mainContent = m.renderContainerSingleView()
	} else if m.isMonitorActive && m.selectedMonitor == 1 {
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
	} else if m.isMonitorActive && m.selectedMonitor == 2 {
		// Vue des conteneurs avec menu contextuel si nécessaire
		containerView := m.renderContainersTable()
		if m.containerMenuState == ContainerMenuVisible {
			// Afficher le menu par-dessus la table
			containerMenu := m.renderContainerMenu()
			// Positionner le menu à droite de la table
			mainContent = lipgloss.JoinHorizontal(lipgloss.Top, containerView, containerMenu)
		} else {
			mainContent = containerView
		}
	} else {
		if m.activeView != -1 {
			m.viewport.SetContent(currentView)
		}
		mainContent = m.viewport.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		home,
		tabs,
		mainContent,
		footer,
	)
}

// ALTERNATIVE
/*

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	if m.err != nil {
		return fmt.Sprintf("Erreur: %v\n", m.err)
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

	// Vérifier d'abord si le menu contextuel est visible (priorité)
	if m.containerMenuState == ContainerMenuVisible {
		mainContent = m.renderContainerMenu()
	} else if m.containerViewState == ContainerViewSingle {
		mainContent = m.renderContainerSingleView()
	} else if m.isMonitorActive && m.selectedMonitor == 1 {
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
	} else if m.isMonitorActive && m.selectedMonitor == 2 {
		// Vue des conteneurs avec menu contextuel si nécessaire
		if m.containerMenuState == ContainerMenuVisible {
			// Afficher uniquement le menu contextuel (remplace le contenu)
			mainContent = m.renderContainerMenu()
		} else {
			mainContent = m.renderContainersTable()
		}
	} else {
		if m.activeView != -1 {
			m.viewport.SetContent(currentView)
		}
		mainContent = m.viewport.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		home,
		tabs,
		mainContent,
		footer,
	)
}

 */