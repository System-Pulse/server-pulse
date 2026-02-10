package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moby/moby/api/types/container"
)

func (dm *DockerManager) RestartContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timeout := int((10 * time.Second).Nanoseconds())
	if err := dm.Cli.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to restart container %s: %w", containerID, err)
	}

	return nil
}

func (dm *DockerManager) StartContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := dm.Cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %w", containerID, err)
	}

	return nil
}

func (dm *DockerManager) StopContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timeout := int((10 * time.Second).Nanoseconds())
	if err := dm.Cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	return nil
}

func (dm *DockerManager) PauseContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dm.Cli.ContainerPause(ctx, containerID); err != nil {
		return fmt.Errorf("failed to pause container %s: %w", containerID, err)
	}

	return nil
}

func (dm *DockerManager) UnpauseContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dm.Cli.ContainerUnpause(ctx, containerID); err != nil {
		return fmt.Errorf("failed to unpause container %s: %w", containerID, err)
	}

	return nil
}

func (dm *DockerManager) DeleteContainer(containerID string, force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := container.RemoveOptions{
		Force: force,
	}

	if err := dm.Cli.ContainerRemove(ctx, containerID, options); err != nil {
		return fmt.Errorf("failed to delete container %s: %w", containerID, err)
	}

	return nil
}

func (dm *DockerManager) GetContainerLogs(containerID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
		Follow:     false,
		Tail:       "all",
	}

	logs, err := dm.Cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for container %s: %w", containerID, err)
	}
	defer logs.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(logs)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 8 && (line[0] == 1 || line[0] == 2) {
			line = line[8:]
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading logs: %w", err)
	}

	return result.String(), nil
}

func (dm *DockerManager) StreamContainerLogs(containerID string) (chan string, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	containerJSON, err := dm.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	if containerJSON.State.Status != "running" {
		cancel()
		return nil, nil, fmt.Errorf("container is not running (status: %s), streaming not available", containerJSON.State.Status)
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
		Follow:     true,
		Tail:       "100",
	}

	logsReader, err := dm.Cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to stream logs: %w", err)
	}

	logChan := make(chan string, 50)

	go func() {
		defer close(logChan)
		defer logsReader.Close()

		scanner := bufio.NewScanner(logsReader)

		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				line := scanner.Text()

				cleanLine := line
				if len(line) > 8 && (line[0] == 1 || line[0] == 2) {
					cleanLine = line[8:]
				}

				select {
				case logChan <- cleanLine:
				case <-ctx.Done():
					return
				case <-time.After(1 * time.Second):
					continue
				}
			}
		}

		if err := scanner.Err(); err != nil && !strings.Contains(err.Error(), "context canceled") {
			select {
			case logChan <- fmt.Sprintf("ERROR: %v", err):
			case <-ctx.Done():
			}
		}
	}()

	return logChan, cancel, nil
}

func (dm *DockerManager) StartLogsStreamCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		logChan, cancelFunc, err := dm.StreamContainerLogs(containerID)
		if err != nil {
			return ContainerLogsMsg{
				ContainerID: containerID,
				Logs:        "",
				Error:       fmt.Errorf("streaming unavailable: %w", err),
			}
		}

		return ContainerLogsStreamMsg{
			ContainerID: containerID,
			LogChan:     logChan,
			CancelFunc:  cancelFunc,
		}
	}
}

func (dm *DockerManager) StopLogsStreamCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		return ContainerLogsStopMsg{
			ContainerID: containerID,
		}
	}
}

func (dm *DockerManager) GetContainerStatus(containerID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containerJSON, err := dm.Cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	return containerJSON.State.Status, nil
}

func (dm *DockerManager) IsContainerRunning(containerID string) (bool, error) {
	status, err := dm.GetContainerStatus(containerID)
	if err != nil {
		return false, err
	}

	return status == "running", nil
}

func (dm *DockerManager) IsContainerPaused(containerID string) (bool, error) {
	status, err := dm.GetContainerStatus(containerID)
	if err != nil {
		return false, err
	}

	return status == "paused", nil
}

func (dm *DockerManager) ToggleContainerState(containerID string) error {
	isRunning, err := dm.IsContainerRunning(containerID)
	if err != nil {
		return err
	}

	if isRunning {
		return dm.StopContainer(containerID)
	} else {
		return dm.StartContainer(containerID)
	}
}

func (dm *DockerManager) ToggleContainerPause(containerID string) error {
	isPaused, err := dm.IsContainerPaused(containerID)
	if err != nil {
		return err
	}

	if isPaused {
		return dm.UnpauseContainer(containerID)
	} else {
		isRunning, err := dm.IsContainerRunning(containerID)
		if err != nil {
			return err
		}

		if !isRunning {
			return fmt.Errorf("cannot pause container %s: container is not running", containerID)
		}

		return dm.PauseContainer(containerID)
	}
}

func isValidContainerID(id string) bool {
	if len(id) < 1 || len(id) > 64 {
		return false
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func (dm *DockerManager) ExecInteractiveShellAlternative(containerID string) error {
	if !isValidContainerID(containerID) {
		return fmt.Errorf("invalid container ID: %s", containerID)
	}

	fmt.Print("\033[?1049l")
	fmt.Print("\033[2J\033[H")

	fmt.Println("Type 'exit' to return to Server-Pulse")

	cmd := exec.Command("docker", "exec", "-it", containerID, "sh", "-c", "command -v bash >/dev/null 2>&1 && exec bash || exec sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	dm.forceTerminalResetSimple()

	if err != nil {
		fmt.Printf("Shell session ended with error: %v\n", err)
	}

	// dm.waitSimple()

	return err
}

func (dm *DockerManager) forceTerminalResetSimple() {
	// Reset complet du terminal
	fmt.Print("\033c") // Full terminal reset
	time.Sleep(100 * time.Millisecond)
	fmt.Print("\033[2J\033[H") // Clear and home
}

func (dm *DockerManager) GetContainerStatsStream(containerID string) (chan ContainerStatsMsg, context.CancelFunc, error) {
	statsChan := make(chan ContainerStatsMsg, 1)
	ctx, cancel := context.WithCancel(context.Background())

	response, err := dm.Cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		cancel()
		close(statsChan)
		return nil, nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	go func() {
		defer response.Body.Close()
		defer close(statsChan)

		decoder := json.NewDecoder(response.Body)
		var lastCPUStats *CPUStats
		var lastSystemCPUUsage uint64
		var lastPerCPUUsage []uint64

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var stats StatsJSON
			if err := decoder.Decode(&stats); err != nil {
				if err == io.EOF || ctx.Err() != nil {
					return
				}
				continue
			}

			cpuPercent := 0.0
			perCpuPercents := make([]float64, len(stats.CPUStats.CPUUsage.PercpuUsage))

			if lastCPUStats != nil && lastSystemCPUUsage > 0 {
				cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - lastCPUStats.CPUUsage.TotalUsage)
				systemDelta := float64(stats.CPUStats.SystemUsage - lastSystemCPUUsage)

				// Handle division by zero
				if systemDelta == 0 {
					// Use previous values if available
					cpuPercent = (cpuDelta / 1) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
				} else {
					numberCpus := float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
					if numberCpus == 0 {
						numberCpus = 1 // Fallback to 1 CPU
					}
					cpuPercent = (cpuDelta / systemDelta) * numberCpus * 100.0
				}

				// Calculate per-core CPU usage
				if len(stats.CPUStats.CPUUsage.PercpuUsage) > 0 && len(lastPerCPUUsage) == len(stats.CPUStats.CPUUsage.PercpuUsage) {
					for i, usage := range stats.CPUStats.CPUUsage.PercpuUsage {
						if i < len(lastPerCPUUsage) {
							perCpuDelta := float64(usage) - float64(lastPerCPUUsage[i])
							perCpuPercent := (perCpuDelta / systemDelta) * 100.0
							perCpuPercents[i] = perCpuPercent
						}
					}
				}

				if cpuPercent > 100 {
					cpuPercent = 100
				}
			}

			// Update last values for next iteration
			lastPerCPUUsage = make([]uint64, len(stats.CPUStats.CPUUsage.PercpuUsage))
			copy(lastPerCPUUsage, stats.CPUStats.CPUUsage.PercpuUsage)

			memoryPercent := 0.0
			if stats.MemoryStats.Limit > 0 {
				memoryPercent = (float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)) * 100.0
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

			select {
			case statsChan <- ContainerStatsMsg{
				ContainerID:    containerID,
				CPUPercent:     cpuPercent,
				PerCPUPercents: perCpuPercents,
				MemPercent:     memoryPercent,
				MemUsage:       stats.MemoryStats.Usage,
				MemLimit:       stats.MemoryStats.Limit,
				NetRX:          networkRx,
				NetTX:          networkTx,
				DiskUsage:      blockRead + blockWrite,
			}:
			case <-ctx.Done():
				return
			}

			lastCPUStats = &stats.CPUStats
			lastSystemCPUUsage = stats.CPUStats.SystemUsage

			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				return
			}
		}
	}()

	return statsChan, cancel, nil
}
