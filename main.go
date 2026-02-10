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
var currentModel tea.Model

func main() {
	if ok, err := utils.CheckDockerPermissions(); !ok {
		fmt.Println(err)
		os.Exit(1)
	}
	defer panicExit()

	// Initialize docker manager globally (non-fatal if Docker is unavailable)
	var err error
	dockerManager, err = app.NewDockerManager()
	if err != nil {
		// Docker is unavailable but the app can still run without container management
		dockerManager = nil
	}

	for {
		if shouldExit := runTUI(); shouldExit {
			break
		}
	}
}

func runTUI() bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			fmt.Println("\nRestoring terminal after panic...")
			time.Sleep(2 * time.Second)
		}
	}()

	lipgloss.SetHasDarkBackground(true)

	var m tea.Model
	if currentModel != nil {
		if model, ok := currentModel.(widgets.Model); ok {
			model.ClearPendingShellExec()
			m = model
		} else {
			m = widgets.InitialModelWithManager(dockerManager)
		}
	} else {
		m = widgets.InitialModelWithManager(dockerManager)
	}

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

	currentModel = finalModel

	// Check if we need to execute a shell
	if model, ok := finalModel.(widgets.Model); ok {
		if shellRequest := model.GetPendingShellExec(); shellRequest != nil && dockerManager != nil {
			time.Sleep(100 * time.Millisecond)

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
