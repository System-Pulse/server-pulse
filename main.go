package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/System-Pulse/server-pulse/system/app"
	"github.com/System-Pulse/server-pulse/utils"
	widgets "github.com/System-Pulse/server-pulse/widgets"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var dockerManager *app.DockerManager

func main() {
	if ok, err := utils.CheckDockerPermissions(); !ok {
		fmt.Println(err)
		os.Exit(1)
	}
	defer panicExit()

	// Initialize docker manager globally
	var err error
	dockerManager, err = app.NewDockerManager()
	if err != nil {
		fmt.Printf("Failed to initialize Docker manager: %v\n", err)
		os.Exit(1)
	}

	// Main application loop
	for {
		if shouldExit := runTUI(); shouldExit {
			break
		}
	}
}

func runTUI() bool {
	lipgloss.SetHasDarkBackground(true)

	m := widgets.InitialModelWithManager(dockerManager)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithMouseAllMotion(),
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Erreur: %v", err)
		os.Exit(1)
	}

	// Check if we need to execute a shell
	if model, ok := finalModel.(widgets.Model); ok {
		if shellRequest := model.GetPendingShellExec(); shellRequest != nil {
			// Attendre un peu avant d'exécuter le shell pour stabiliser
			time.Sleep(100 * time.Millisecond)

			// Execute shell outside TUI
			err := dockerManager.ExecInteractiveShellAlternative(shellRequest.ContainerID)
			if err != nil {
				fmt.Printf("Shell execution failed: %v\n", err)
				fmt.Println("Press Enter to continue...")
				bufio.NewScanner(os.Stdin).Scan()
			}

			time.Sleep(300 * time.Millisecond)

			// Return false to restart TUI
			return false
		}
		// Check if user wants to quit
		return model.ShouldQuit()
	}

	return true
}

func panicExit() {
	if r := recover(); r != nil {
		log.Println("shutting down")
		panic(r)
	}
}
