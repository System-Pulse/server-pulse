package app

import (
	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func (dm *DockerManager) UpdateApp() tea.Cmd {
	return func() tea.Msg {
		cont, err := dm.RefreshContainers()
		if err != nil {
			return utils.ErrMsg(err)
		}
		return ContainerMsg(cont)
	}
}

// ContainerStatsMsg contient les statistiques en temps réel d'un conteneur
type ContainerStatsMsg struct {
	ContainerID string
	CPUPercent  float64
	MemPercent  float64
	MemUsage    uint64
	MemLimit    uint64
	NetRX       uint64
	NetTX       uint64
	DiskUsage   uint64
}

// GetContainerStats récupère les statistiques en temps réel d'un conteneur
// Version simplifiée avec données simulées pour l'instant
func (dm *DockerManager) GetContainerStats(containerID string) tea.Cmd {
	return func() tea.Msg {
		// Pour l'instant, retourner des statistiques simulées
		// TODO: Implémenter la vraie collecte de statistiques Docker
		return ContainerStatsMsg{
			ContainerID: containerID,
			CPUPercent:  35.5 + float64(len(containerID)%30), // Simulé
			MemPercent:  45.2 + float64(len(containerID)%25), // Simulé
			MemUsage:    1024 * 1024 * 512,                   // 512MB simulé
			MemLimit:    1024 * 1024 * 1024,                  // 1GB simulé
			NetRX:       1024 * 1024 * 10,                    // 10MB simulé
			NetTX:       1024 * 1024 * 5,                     // 5MB simulé
			DiskUsage:   1024 * 1024 * 100,                   // 100MB simulé
		}
	}
}
