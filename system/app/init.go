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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	containers, err := dm.Cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []Container
	var skipped int

	for _, cont := range containers {
		containerJSON, err := dm.Cli.ContainerInspect(ctx, cont.ID)
		if err != nil {
			skipped++
			continue
		}

		projectName := cont.Labels["com.docker.compose.project"]
		if projectName == "" {
			projectName = "N/A"
		}

		containerName := "N/A"
		if len(cont.Names) > 0 {
			containerName = strings.TrimPrefix(cont.Names[0], "/")
			containerName = strings.TrimPrefix(containerName, projectName+"-")
		}

		status := strings.Split(cont.Status, " ")[0]
		state := cont.State

		var portsInfo []string
		for _, p := range cont.Ports {
			if p.PublicPort > 0 {
				portsInfo = append(portsInfo, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				portsInfo = append(portsInfo, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}
		portsStr := strings.Join(portsInfo, ", ")
		if portsStr == "" {
			portsStr = "N/A"
		}

		health := "N/A"
		if containerJSON.State.Health != nil {
			health = string(containerJSON.State.Health.Status)
		} else {
			switch state {
			case "running":
				health = "running"
			case "exited":
				health = "exited"
			case "paused":
				health = "paused"
			case "created":
				health = "created"
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
			PortsStr:  portsStr,
			Health:    health,
		}
		result = append(result, c)
	}

	if skipped > 0 {
		return result, fmt.Errorf("%d container(s) could not be inspected", skipped)
	}

	return result, nil
}

func (dm *DockerManager) GetContainerDetails(containerID string) (*ContainerDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	containerJSON, err := dm.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	statsResponse, err := dm.Cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer statsResponse.Body.Close()

	var stats StatsJSON
	if err := json.NewDecoder(statsResponse.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	var cpuPercent float64
	if stats.PreCPUStats.CPUUsage.TotalUsage > 0 && stats.CPUStats.SystemUsage > stats.PreCPUStats.SystemUsage {
		cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

		// Handle division by zero
		if systemDelta == 0 {
			// Use a small value to avoid division by zero
			systemDelta = 1
		}

		numberCpus := float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		if numberCpus == 0 {
			numberCpus = 1 // Fallback to 1 CPU
		}

		cpuPercent = (cpuDelta / systemDelta) * numberCpus * 100.0

		// Cap CPU percentage at 100%
		if cpuPercent > 100 {
			cpuPercent = 100
		}
	}

	memoryUsage := stats.MemoryStats.Usage
	memoryLimit := stats.MemoryStats.Limit
	memoryPercent := 0.0
	if memoryLimit > 0 {
		memoryPercent = (float64(memoryUsage) / float64(memoryLimit)) * 100.0
	}

	var networkRx, networkTx uint64
	if stats.Networks != nil {
		for _, network := range stats.Networks {
			networkRx += network.RxBytes
			networkTx += network.TxBytes
		}
	}

	var blockRead, blockWrite uint64
	for _, ioStat := range stats.BlkioStats.IoServiceBytesRecursive {
		switch ioStat.Op {
		case "Read":
			blockRead += ioStat.Value
		case "Write":
			blockWrite += ioStat.Value
		}
	}

	uptime := "N/A"
	if containerJSON.State.StartedAt != "" {
		startTime, err := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
		if err == nil {
			uptime = formatDuration(time.Since(startTime))
		}
	}

	ipAddress := "N/A"
	// gateway := "N/A"
	if containerJSON.NetworkSettings != nil && len(containerJSON.NetworkSettings.Networks) > 0 {
		for netName, network := range containerJSON.NetworkSettings.Networks {
			if ipAddress == "N/A" || netName == "bridge" {
				ipAddress = network.IPAddress
				// gateway = network.Gateway
			}
		}
	}

	// Health check
	healthCheck := "N/A"
	if containerJSON.State.Health != nil {
		healthCheck = string(containerJSON.State.Health.Status)
	}

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

	var ipAddresses []string
	var gateways []string
	if containerJSON.NetworkSettings != nil && len(containerJSON.NetworkSettings.Networks) > 0 {
		for netName, network := range containerJSON.NetworkSettings.Networks {
			if network.IPAddress != "" {
				ipAddresses = append(ipAddresses, fmt.Sprintf("%s: %s", netName, network.IPAddress))
			}
			if network.Gateway != "" {
				gateways = append(gateways, fmt.Sprintf("%s: %s", netName, network.Gateway))
			}
		}
	}

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
		Environment:     containerJSON.Config.Env,
		IPAddress:       strings.Join(ipAddresses, ", "),
		Gateway:         strings.Join(gateways, ", "),
		HealthCheck:     healthCheck,
		Uptime:          uptime,
		Ports:           ports,
		NetworkSettings: containerJSON.NetworkSettings,
		HostConfig:      containerJSON.HostConfig,
		State:           &containerJSON.State,
	}

	return details, nil
}

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
