package widgets

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderMonitor() string {
	var currentView string
	switch m.Ui.SelectedMonitor {
	case 0:
		currentView = m.renderSystem()
	case 1:
		currentView = m.renderProcesses()
	case 2:
		currentView = m.renderContainers()
	}
	return currentView
}

func (m Model) renderContainers() string {
	p := "Search a container..."
	return m.renderTable(m.Monitor.Container, p)
}

func (m Model) renderProcesses() string {
	p := "Search a process..."
	return m.renderTable(m.Monitor.ProcessTable, p)
}

func (m Model) renderSystem() string {
	doc := strings.Builder{}
	cpuInfo := fmt.Sprintf("CPU: %s %.1f%% | Load: %.2f, %.2f, %.2f", m.Monitor.CpuProgress.View(), m.Monitor.Cpu.Usage, m.Monitor.Cpu.LoadAvg1, m.Monitor.Cpu.LoadAvg5, m.Monitor.Cpu.LoadAvg15)
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("CPU"))
	doc.WriteString("\n")
	doc.WriteString(cpuInfo)
	doc.WriteString("\n\n")
	memInfo := fmt.Sprintf("RAM: %s %.1f%% | Total: %s | Used: %s | Free: %s", m.Monitor.MemProgress.View(), m.Monitor.Memory.Usage, utils.FormatBytes(m.Monitor.Memory.Total), utils.FormatBytes(m.Monitor.Memory.Used), utils.FormatBytes(m.Monitor.Memory.Free))
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Memory"))
	doc.WriteString("\n")
	doc.WriteString(memInfo)
	doc.WriteString("\n")
	swapInfo := fmt.Sprintf("SWP: %s %.1f%% | Total: %s | Used: %s | Free: %s", m.Monitor.SwapProgress.View(), m.Monitor.Memory.SwapUsage, utils.FormatBytes(m.Monitor.Memory.SwapTotal), utils.FormatBytes(m.Monitor.Memory.SwapUsed), utils.FormatBytes(m.Monitor.Memory.SwapFree))
	doc.WriteString(swapInfo)
	doc.WriteString("\n\n")
	doc.WriteString(lipgloss.NewStyle().Bold(true).Render("Disks"))
	doc.WriteString("\n")
	for _, disk := range m.Monitor.Disks {
		if disk.Total > 0 {
			if p, ok := m.Monitor.DiskProgress[disk.Mountpoint]; ok {
				diskInfo := fmt.Sprintf("%-10s %s %.1f%% (%s/%s)", utils.Ellipsis(disk.Mountpoint, 10), p.View(), disk.Usage, utils.FormatBytes(disk.Used), utils.FormatBytes(disk.Total))
				doc.WriteString(diskInfo)
				doc.WriteString("\n")
			}
		}
	}
	return doc.String()
}
