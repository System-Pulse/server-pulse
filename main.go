package main

import (
	"fmt"
	"log"
	"os"

	"github.com/System-Pulse/server-pulse/utils"
	app "github.com/System-Pulse/server-pulse/widgets"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	if ok, err := utils.CheckDockerPermissions(); !ok {
		fmt.Println(err)
		os.Exit(1)
	}
	defer panicExit()
	
	lipgloss.SetHasDarkBackground(true)

	m := app.InitialModel()

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur: %v", err)
		os.Exit(1)
	}
}

func panicExit() {
	if r := recover(); r != nil {
		log.Println("shutting down")
		panic(r)
	}
}
