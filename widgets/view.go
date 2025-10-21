package widgets

import (
	"fmt"

	v "github.com/System-Pulse/server-pulse/widgets/vars"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {

	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}
	if !m.Ui.Ready {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render("Terminal too small!\nMinimum size: (min: 80x20)\nCurrent: " +
				fmt.Sprintf("%dx%d", m.Ui.Width, m.Ui.Height))
	}

	header := m.renderHeader()
	nav := m.renderCurrentNav()
	mainContent := m.renderMainContent()
	footer := m.renderFooter()
	if m.Monitor.ContainerMenuState == v.ContainerMenuVisible {
		mainContent = m.renderContainerMenu()
	} else if m.ConfirmationVisible {
		mainContent = m.renderConfirmationDialog()
	}

	baseHeaderView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		nav,
	)

	baseView := lipgloss.JoinVertical(lipgloss.Left,
		baseHeaderView,
		mainContent,
		footer,
	)

	return baseView
}
