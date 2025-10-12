package performance

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/utils"
	"github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/process"
)

// GetIOMetrics collects comprehensive I/O performance metrics
func GetIOMetrics() tea.Cmd {
	return func() tea.Msg {
		metrics, err := collectIOMetrics()
		return IOMetricsMsg{
			Metrics: metrics,
			Error:   err,
		}
	}
}

func collectIOMetrics() (*model.IOMetrics, error) {
	metrics := &model.IOMetrics{
		LastUpdate: time.Now(),
	}

	// Collect disk I/O statistics
	disks, err := getDiskIOStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk I/O stats: %v", err)
	}
	metrics.Disks = disks

	// Calculate totals
	for _, disk := range disks {
		metrics.TotalReadIOPS += disk.ReadIOPS
		metrics.TotalWriteIOPS += disk.WriteIOPS
		metrics.TotalReadBytes += disk.ReadBytes
		metrics.TotalWriteBytes += disk.WriteBytes
	}

	// Collect top I/O processes
	processes, err := getTopIOProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to get process I/O stats: %v", err)
	}
	metrics.TopProcesses = processes

	// Calculate average latency (simplified)
	if len(disks) > 0 {
		totalTime := uint64(0)
		totalOps := uint64(0)
		for _, disk := range disks {
			totalTime += disk.ReadTime + disk.WriteTime
			totalOps += disk.ReadIOPS + disk.WriteIOPS
		}
		if totalOps > 0 {
			metrics.AverageLatency = float64(totalTime) / float64(totalOps)
		}
	}

	return metrics, nil
}

func getDiskIOStats() ([]model.DiskIOInfo, error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var disks []model.DiskIOInfo
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		// Skip if not enough fields (should have at least 14 fields)
		if len(fields) < 14 {
			continue
		}

		device := fields[2]

		// Skip loop devices, ram devices, and other non-physical devices
		if strings.HasPrefix(device, "loop") ||
			strings.HasPrefix(device, "ram") ||
			strings.HasPrefix(device, "fd") ||
			strings.HasPrefix(device, "sr") {
			continue
		}

		// Parse disk statistics
		// Fields: major minor name reads reads_merged reads_sectors reads_time writes writes_merged writes_sectors writes_time in_progress io_time weighted_io_time
		reads, _ := strconv.ParseUint(fields[3], 10, 64)
		readTime, _ := strconv.ParseUint(fields[6], 10, 64)
		writes, _ := strconv.ParseUint(fields[7], 10, 64)
		writeTime, _ := strconv.ParseUint(fields[10], 10, 64)
		inProgress, _ := strconv.ParseUint(fields[11], 10, 64)
		ioTime, _ := strconv.ParseUint(fields[12], 10, 64)

		// Convert sectors to bytes (assuming 512-byte sectors)
		readSectors, _ := strconv.ParseUint(fields[5], 10, 64)
		writeSectors, _ := strconv.ParseUint(fields[9], 10, 64)
		readBytes := readSectors * 512
		writeBytes := writeSectors * 512

		// Calculate utilization percentage
		utilization := float64(0)
		if ioTime > 0 {
			utilization = float64(ioTime) / 10.0 // Convert from 10ms units to percentage
		}

		disk := model.DiskIOInfo{
			Device:      device,
			ReadIOPS:    reads,
			WriteIOPS:   writes,
			ReadBytes:   readBytes,
			WriteBytes:  writeBytes,
			ReadTime:    readTime,
			WriteTime:   writeTime,
			QueueDepth:  inProgress,
			Utilization: utilization,
		}

		disks = append(disks, disk)
	}

	// Sort by total I/O activity
	sort.Slice(disks, func(i, j int) bool {
		totalI := disks[i].ReadIOPS + disks[i].WriteIOPS
		totalJ := disks[j].ReadIOPS + disks[j].WriteIOPS
		return totalI > totalJ
	})

	return disks, scanner.Err()
}

func getTopIOProcesses() ([]model.ProcessIOInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var ioProcesses []model.ProcessIOInfo

	for _, p := range processes {
		// Get I/O counters
		ioCounters, err := p.IOCounters()
		if err != nil {
			continue // Skip processes we can't access
		}

		// Get process name
		name, err := p.Name()
		if err != nil {
			name = "unknown"
		}

		ioProcess := model.ProcessIOInfo{
			PID:        p.Pid,
			Command:    name,
			ReadIOPS:   ioCounters.ReadCount,
			WriteIOPS:  ioCounters.WriteCount,
			ReadBytes:  ioCounters.ReadBytes,
			WriteBytes: ioCounters.WriteBytes,
		}

		// Only include processes with some I/O activity
		if ioProcess.ReadIOPS > 0 || ioProcess.WriteIOPS > 0 {
			ioProcesses = append(ioProcesses, ioProcess)
		}
	}

	// Sort by total I/O bytes
	sort.Slice(ioProcesses, func(i, j int) bool {
		totalI := ioProcesses[i].ReadBytes + ioProcesses[i].WriteBytes
		totalJ := ioProcesses[j].ReadBytes + ioProcesses[j].WriteBytes
		return totalI > totalJ
	})

	// Limit to top 10
	if len(ioProcesses) > 10 {
		ioProcesses = ioProcesses[:10]
	}

	return ioProcesses, nil
}

func RenderInputOutput() string {
	return vars.CardStyle.Render("⏳ Loading I/O Performance Metrics...")
}

func RenderInputOutputWithData(metrics *model.IOMetrics, loading bool) string {
	if loading {
		return vars.CardStyle.Render("⏳ Loading I/O Performance Metrics...")
	}

	if metrics == nil {
		return vars.CardStyle.Render("⏳ Loading I/O Performance Metrics...")
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(vars.AccentColor).MarginBottom(1)
	b.WriteString(titleStyle.Render("─ I/O PERFORMANCE ANALYSIS ─"))
	b.WriteString("\n")

	// Summary section
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("📊 I/O SUMMARY"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("│ Total Read IOPS:  %s\n", formatNumber(metrics.TotalReadIOPS)))
	b.WriteString(fmt.Sprintf("│ Total Write IOPS: %s\n", formatNumber(metrics.TotalWriteIOPS)))
	b.WriteString(fmt.Sprintf("│ Total Read:       %s\n", utils.FormatBytes(metrics.TotalReadBytes)))
	b.WriteString(fmt.Sprintf("│ Total Write:      %s\n", utils.FormatBytes(metrics.TotalWriteBytes)))
	b.WriteString(fmt.Sprintf("│ Avg Latency:      %.2f ms\n", metrics.AverageLatency))
	b.WriteString(fmt.Sprintf("│ Last Update:      %s\n", metrics.LastUpdate.Format("15:04:05")))
	b.WriteString("\n")

	// Disk I/O table
	if len(metrics.Disks) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("💾 DISK I/O PERFORMANCE"))
		b.WriteString("\n")

		// Create table
		columns := []table.Column{
			{Title: "Device", Width: 8},
			{Title: "Read IOPS", Width: 10},
			{Title: "Write IOPS", Width: 11},
			{Title: "Read MB/s", Width: 10},
			{Title: "Write MB/s", Width: 11},
			{Title: "Util%", Width: 6},
			{Title: "Queue", Width: 6},
		}

		var rows []table.Row
		for _, disk := range metrics.Disks {
			readMBps := float64(disk.ReadBytes) / (1024 * 1024)
			writeMBps := float64(disk.WriteBytes) / (1024 * 1024)

			rows = append(rows, table.Row{
				disk.Device,
				formatNumber(disk.ReadIOPS),
				formatNumber(disk.WriteIOPS),
				fmt.Sprintf("%.1f", readMBps),
				fmt.Sprintf("%.1f", writeMBps),
				fmt.Sprintf("%.1f", disk.Utilization),
				formatNumber(disk.QueueDepth),
			})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(false),
			table.WithHeight(min(8, len(rows)+1)),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		b.WriteString(t.View())
		b.WriteString("\n\n")
	}

	// Top I/O Processes
	if len(metrics.TopProcesses) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("🔥 TOP I/O PROCESSES"))
		b.WriteString("\n")

		columns := []table.Column{
			{Title: "PID", Width: 6},
			{Title: "Command", Width: 20},
			{Title: "Read IOPS", Width: 10},
			{Title: "Write IOPS", Width: 11},
			{Title: "Read MB", Width: 10},
			{Title: "Write MB", Width: 11},
		}

		var rows []table.Row
		for _, proc := range metrics.TopProcesses {
			readMB := float64(proc.ReadBytes) / (1024 * 1024)
			writeMB := float64(proc.WriteBytes) / (1024 * 1024)

			rows = append(rows, table.Row{
				fmt.Sprintf("%d", proc.PID),
				utils.Ellipsis(proc.Command, 18),
				formatNumber(proc.ReadIOPS),
				formatNumber(proc.WriteIOPS),
				fmt.Sprintf("%.1f", readMB),
				fmt.Sprintf("%.1f", writeMB),
			})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(false),
			table.WithHeight(min(6, len(rows)+1)),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		b.WriteString(t.View())
		b.WriteString("\n\n")
	}

	// I/O Health Indicators
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("⚡ I/O HEALTH INDICATORS"))
	b.WriteString("\n")

	// Check for high utilization disks
	highUtilDisks := 0
	for _, disk := range metrics.Disks {
		if disk.Utilization > 80 {
			highUtilDisks++
		}
	}

	if highUtilDisks > 0 {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		b.WriteString(warningStyle.Render(fmt.Sprintf("⚠️  %d disk(s) with high utilization (>80%%)\n", highUtilDisks)))
	} else {
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		b.WriteString(okStyle.Render("✅ All disks operating within normal utilization\n"))
	}

	// Check for high queue depth
	highQueueDisks := 0
	for _, disk := range metrics.Disks {
		if disk.QueueDepth > 10 {
			highQueueDisks++
		}
	}

	if highQueueDisks > 0 {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		b.WriteString(warningStyle.Render(fmt.Sprintf("⚠️  %d disk(s) with high queue depth (>10)\n", highQueueDisks)))
	} else {
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		b.WriteString(okStyle.Render("✅ All disks have normal queue depth\n"))
	}

	// Check for high latency
	if metrics.AverageLatency > 10 {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		b.WriteString(warningStyle.Render(fmt.Sprintf("⚠️  High average I/O latency: %.2f ms\n", metrics.AverageLatency)))
	} else {
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		b.WriteString(okStyle.Render(fmt.Sprintf("✅ Normal I/O latency: %.2f ms\n", metrics.AverageLatency)))
	}

	return vars.CardStyle.Render(b.String())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
