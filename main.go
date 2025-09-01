package main

import (
	"fmt"
	"os"

	app "github.com/System-Pulse/server-pulse/widgets"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Configuration des styles globaux
	lipgloss.SetHasDarkBackground(true)

	// Le modèle est initialisé via la fonction InitialModel
	m := app.InitialModel()

	p := tea.NewProgram(
		m, // Passez le modèle initialisé
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur: %v", err)
		os.Exit(1)
	}
}
