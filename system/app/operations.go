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

	"github.com/System-Pulse/server-pulse/utils"
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

func (dm *DockerManager) GetContainerLogsStream(containerID string, since string, follow bool) (io.ReadCloser, error) {
	ctx := context.Background()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     follow,
		Since:      since,
		Tail:       "100",
	}

	return dm.Cli.ContainerLogs(ctx, containerID, options)
}

func (dm *DockerManager) GetContainerLogs(containerID string, tail string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       tail, 
	}

	logs, err := dm.Cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for container %s: %w", containerID, err)
	}
	defer logs.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(logs)
	lineCount := 0
	maxLines := 500 

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		if len(line) > 8 && (line[0] == 1 || line[0] == 2) {
			line = line[8:]
		}
		result.WriteString(line)
		result.WriteString("\n")
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading logs: %w", err)
	}

	return result.String(), nil
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

func (dm *DockerManager) ExecInteractiveShellAlternative(containerID string) error {
	fmt.Print("\033[?1049l")
	fmt.Print("\033[2J\033[H")

	fmt.Println("Type 'exit' to return to Server-Pulse")

	cmd := exec.Command("docker", "exec", "-it", containerID, "sh", "-c", "command -v bash >/dev/null 2>&1 && exec bash || exec sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	dm.forceTerminalResetSimple()

	fmt.Println("=== Container shell session ended ===")
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

/*func (dm *DockerManager) waitSimple() {
	fmt.Print("Press Enter to return to Server-Pulse...")
	cmd := exec.Command("bash", "-c", "read -r")
	cmd.Stdin = os.Stdin
	cmd.Run()
	fmt.Print("\033[2J\033[H")
}*/

func (dm *DockerManager) GetContainerStatsStream(containerID string) (chan ContainerStatsMsg, error) {
	statsChan := make(chan ContainerStatsMsg)
	ctx := context.Background()

	response, err := dm.Cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		close(statsChan)
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	go func() {
		defer response.Body.Close()
		defer close(statsChan)

		decoder := json.NewDecoder(response.Body)
		var lastCPUStats *CPUStats
		var lastSystemCPUUsage uint64

		for {
			var stats StatsJSON
			if err := decoder.Decode(&stats); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			cpuPercent := 0.0
			if lastCPUStats != nil && lastSystemCPUUsage > 0 {
				cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - lastCPUStats.CPUUsage.TotalUsage)
				systemDelta := float64(stats.CPUStats.SystemUsage - lastSystemCPUUsage)

				if systemDelta > 0 && cpuDelta > 0 {
					cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
					if cpuPercent > 100 {
						cpuPercent = 100
					}
				}
			}

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

			statsChan <- ContainerStatsMsg{
				ContainerID: containerID,
				CPUPercent:  cpuPercent,
				MemPercent:  memoryPercent,
				MemUsage:    stats.MemoryStats.Usage,
				MemLimit:    stats.MemoryStats.Limit,
				NetRX:       networkRx,
				NetTX:       networkTx,
				DiskUsage:   blockRead + blockWrite,
			}

			lastCPUStats = &stats.CPUStats
			lastSystemCPUUsage = stats.CPUStats.SystemUsage

			time.Sleep(2 * time.Second) 
		}
	}()

	return statsChan, nil
}

func (dm *DockerManager) GetContainerStatsCmd(containerID string) tea.Cmd {
	return func() tea.Msg {
		statsChan, err := dm.GetContainerStatsStream(containerID)
		if err != nil {
			return utils.ErrMsg(err)
		}

		return ContainerStatsChanMsg{
			ContainerID: containerID,
			StatsChan:   statsChan,
		}
	}
}
