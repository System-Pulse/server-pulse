package process

import (
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v4/process"
)

func UpdateProcesses() tea.Cmd {
	return func() tea.Msg {
		processes, err := process.Processes()
		if err != nil {
			return utils.ErrMsg(err)
		}

		var processList []ProcessInfo
		for _, p := range processes {
			name, _ := p.Name()
			user, _ := p.Username()
			cpu, _ := p.CPUPercent()
			mem, _ := p.MemoryPercent()

			processList = append(processList, ProcessInfo{
				PID:     p.Pid,
				User:    user,
				CPU:     cpu,
				Mem:     float64(mem),
				Command: name,
			})
		}

		sort.Slice(processList, func(i, j int) bool {
			return processList[i].CPU > processList[j].CPU
		})

		if len(processList) > 50 {
			processList = processList[:50]
		}

		return ProcessMsg(processList)
	}
}

func StopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	// Poll every 200ms for up to 5 seconds instead of blocking for 5s
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	if processExists(pid) {
		return process.Kill()
	}

	return nil
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
