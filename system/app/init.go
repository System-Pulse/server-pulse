package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func NewDockerManager() (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}

	return &DockerManager{Cli: cli}, nil
}

func (dm *DockerManager) RefreshContainers() ([]Container, error) {
	containers, err := dm.Cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []Container

	for _, cont := range containers {
		projectName := cont.Labels["com.docker.compose.project"]
		if projectName == "" {
			projectName = "N/A"
		}

		containerName := "N/A"
		if len(cont.Names) > 0 {
			containerName = strings.TrimPrefix(cont.Names[0], "/")
		}

		status := strings.Split(cont.Status, " ")[0]
		state := cont.State

		// Santé du conteneur
		health := "N/A"
		if cont.Labels["com.docker.compose.service"] != "" {
			// Pour les conteneurs Docker Compose, on peut essayer de déterminer la santé
			switch state {
			case "running":
				health = "healthy"
			case "exited":
				health = "unhealthy"
			case "paused":
				health = "paused"
			default:
				health = state
			}
		}

		c := Container{
			ID:        cont.ID[:12],
			Name:      containerName,
			Status:    status,
			State:     state,
			CreatedAt: time.Unix(cont.Created, 0).Format(time.RFC3339),
			Project:   projectName,
			Image:     cont.Image,
			Command:   cont.Command,
			Ports:     cont.Ports,
			Health:    health,
		}
		result = append(result, c)
	}

	return result, nil
}

func (dm *DockerManager) GetContainerDetails(containerID string) (*ContainerDetails, error) {
	ctx := context.Background()

	// Récupérer les informations de base du conteneur
	containerJSON, err := dm.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Récupérer les statistiques du conteneur
	statsResponse, err := dm.Cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer statsResponse.Body.Close()

	// Parser manuellement les statistiques
	var stats StatsJSON
	if err := json.NewDecoder(statsResponse.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	// Calculer les métriques CPU
	var cpuPercent float64
	if stats.PreCPUStats.CPUUsage.TotalUsage > 0 && stats.CPUStats.SystemUsage > stats.PreCPUStats.SystemUsage {
		cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
		if systemDelta > 0 {
			cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		}
	}

	// Métriques mémoire
	memoryUsage := stats.MemoryStats.Usage
	memoryLimit := stats.MemoryStats.Limit
	memoryPercent := 0.0
	if memoryLimit > 0 {
		memoryPercent = (float64(memoryUsage) / float64(memoryLimit)) * 100.0
	}

	// Métriques réseau
	var networkRx, networkTx uint64
	if stats.Networks != nil {
		for _, network := range stats.Networks {
			networkRx += network.RxBytes
			networkTx += network.TxBytes
		}
	}

	// Métriques disque
	var blockRead, blockWrite uint64
	for _, ioStat := range stats.BlkioStats.IoServiceBytesRecursive {
		switch ioStat.Op {
		case "Read":
			blockRead += ioStat.Value
		case "Write":
			blockWrite += ioStat.Value
		}
	}

	// Calculer l'uptime
	uptime := "N/A"
	if containerJSON.State.StartedAt != "" {
		startTime, err := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
		if err == nil {
			uptime = formatDuration(time.Since(startTime))
		}
	}

	// Récupérer l'IP et la gateway
	ipAddress := "N/A"
	gateway := "N/A"
	if containerJSON.NetworkSettings != nil && len(containerJSON.NetworkSettings.Networks) > 0 {
		for netName, network := range containerJSON.NetworkSettings.Networks {
			if ipAddress == "N/A" || netName == "bridge" {
				ipAddress = network.IPAddress
				gateway = network.Gateway
			}
		}
	}

	// Health check
	healthCheck := "N/A"
	if containerJSON.State.Health != nil {
		healthCheck = string(containerJSON.State.Health.Status)
	}

	// Récupérer les ports exposés
	var ports []PortInfo
	for port, bindings := range containerJSON.NetworkSettings.Ports {
		for _, binding := range bindings {
			publicPort, _ := strconv.ParseUint(binding.HostPort, 10, 16)
			privatePort, _ := strconv.ParseUint(port.Port(), 10, 16)
			
			ports = append(ports, PortInfo{
				PublicPort:  uint16(publicPort),
				PrivatePort: uint16(privatePort),
				Type:        port.Proto(),
				HostIP:      binding.HostIP,
			})
		}
	}

	// Formater la date de création
	createdAt := containerJSON.Created
	if parsedTime, err := time.Parse(time.RFC3339Nano, containerJSON.Created); err == nil {
		createdAt = parsedTime.Format("2006-01-02 15:04:05")
	}

	details := &ContainerDetails{
		Container: Container{
			ID:        containerJSON.ID[:12],
			Name:      strings.TrimPrefix(containerJSON.Name, "/"),
			Image:     containerJSON.Config.Image,
			Status:    containerJSON.State.Status,
			CreatedAt: createdAt,
			Project:   containerJSON.Config.Labels["com.docker.compose.project"],
			Command:   strings.Join(containerJSON.Config.Cmd, " "),
		},
		Stats: ContainerStats{
			CPUPercent:    cpuPercent,
			MemoryUsage:   memoryUsage,
			MemoryLimit:   memoryLimit,
			MemoryPercent: memoryPercent,
			NetworkRx:     networkRx,
			NetworkTx:     networkTx,
			BlockRead:     blockRead,
			BlockWrite:    blockWrite,
		},
		Environment: containerJSON.Config.Env,
		IPAddress:   ipAddress,
		Gateway:     gateway,
		HealthCheck: healthCheck,
		Uptime:      uptime,
		Ports:       ports,
	}

	return details, nil
}


// Fonction utilitaire pour formater la durée
func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
