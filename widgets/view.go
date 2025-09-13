package widgets

import (
	"fmt"

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
			Render("Terminal too small!\nMinimum size: 40x10\nCurrent: " +
				fmt.Sprintf("%dx%d", m.Ui.Width, m.Ui.Height))
	}

	header := m.renderHeader()
	nav := m.renderCurrentNav()
	mainContent := m.renderMainContent()
	footer := m.renderFooter()
	if m.Monitor.ContainerMenuState == ContainerMenuVisible {
		mainContent = m.renderContainerMenu()
	} else if m.ConfirmationVisible {
		mainContent = m.renderConfirmationDialog()
	}

	mainContentStyled := lipgloss.NewStyle().
		Height(m.Ui.ContentHeight).
		Render(mainContent)

	baseView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		nav,
		mainContentStyled,
		footer,
	)

	// if m.ConfirmationVisible {
	// 	dialog := m.renderConfirmationDialog()
	// 	// Place la boîte de dialogue au centre
	// 	return lipgloss.Place(m.Ui.Width, m.Ui.Height, lipgloss.Center, lipgloss.Center, dialog)
	// }

	return baseView
}
