package performance

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/cpu"
)

// CPUStateBreakdown represents detailed CPU state information
type CPUStateBreakdown struct {
	User      float64
	System    float64
	Idle      float64
	IOWait    float64
	IRQ       float64
	SoftIRQ   float64
	Steal     float64
	Nice      float64
	Guest     float64
	GuestNice float64
}

// CPUCoreInfo represents information for a single CPU core
type CPUCoreInfo struct {
	CoreID      int
	Usage       float64
	Frequency   float64
	Temperature float64 // If available
}

// CPUMetrics holds comprehensive CPU performance metrics
type CPUMetrics struct {
	OverallUsage    float64
	StateBreakdown  CPUStateBreakdown
	Cores           []CPUCoreInfo
	ContextSwitches uint64
	Interrupts      uint64
	LoadAverage     [3]float64
	ProcessCount    int
	ThreadCount     int
	LastUpdate      time.Time
}

// GetCPUMetrics collects comprehensive CPU performance metrics
func GetCPUMetrics() tea.Cmd {
	return func() tea.Msg {
		metrics, err := collectCPUMetrics()
		return CPUMetricsMsg{
			Metrics: metrics,
			Error:   err,
		}
	}
}

func collectCPUMetrics() (*model.CPUMetrics, error) {
	metrics := &model.CPUMetrics{
		LastUpdate: time.Now(),
	}

	// Get overall CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	if len(cpuPercent) > 0 {
		metrics.OverallUsage = cpuPercent[0]
	}

	// Get per-core usage
	cpuPercentAll, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get per-core CPU usage: %v", err)
	}

	// Get CPU times for detailed breakdown
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %v", err)
	}

	if len(cpuTimes) > 0 {
		times := cpuTimes[0]
		total := times.User + times.System + times.Idle + times.Nice +
			times.Iowait + times.Irq + times.Softirq + times.Steal +
			times.Guest + times.GuestNice

		if total > 0 {
			metrics.StateBreakdown = model.CPUStateBreakdown{
				User:      (times.User / total) * 100,
				System:    (times.System / total) * 100,
				Idle:      (times.Idle / total) * 100,
				IOWait:    (times.Iowait / total) * 100,
				IRQ:       (times.Irq / total) * 100,
				SoftIRQ:   (times.Softirq / total) * 100,
				Steal:     (times.Steal / total) * 100,
				Nice:      (times.Nice / total) * 100,
				Guest:     (times.Guest / total) * 100,
				GuestNice: (times.GuestNice / total) * 100,
			}
		}
	}

	// Collect per-core information
	for i, usage := range cpuPercentAll {
		coreInfo := model.CPUCoreInfo{
			CoreID: i,
			Usage:  usage,
		}
		metrics.Cores = append(metrics.Cores, coreInfo)
	}

	// Get system statistics
	contextSwitches, err := getStatValue("/proc/stat", "ctxt")
	if err == nil {
		metrics.ContextSwitches = contextSwitches
	}

	interrupts, err := getStatValue("/proc/stat", "intr")
	if err == nil {
		metrics.Interrupts = interrupts
	}

	// Get load average
	loadAvg, err := getLoadAverage()
	if err == nil {
		metrics.LoadAverage = loadAvg
	}

	// Get process and thread counts
	processCount, threadCount, err := getProcessCounts()
	if err == nil {
		metrics.ProcessCount = processCount
		metrics.ThreadCount = threadCount
	}

	return metrics, nil
}

func getLoadAverage() ([3]float64, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return [3]float64{}, err
	}
	defer file.Close()

	var load1, load5, load15 float64
	_, err = fmt.Fscanf(file, "%f %f %f", &load1, &load5, &load15)
	if err != nil {
		return [3]float64{}, err
	}

	return [3]float64{load1, load5, load15}, nil
}

func getProcessCounts() (int, int, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var processes, threads int

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processes ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				processes, _ = strconv.Atoi(fields[1])
			}
		} else if strings.HasPrefix(line, "procs_running ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				threads, _ = strconv.Atoi(fields[1])
			}
		}
	}

	return processes, threads, scanner.Err()
}

func RenderCPU() string {
	return "⏳ Loading CPU Performance Metrics..."
}

// RenderCPUContent generates the CPU content without wrapping in CardStyle
// This allows it to be used in a viewport
func RenderCPUContent(metrics *model.CPUMetrics) string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(vars.AccentColor).MarginBottom(1)
	b.WriteString(titleStyle.Render("═══════════════════════════════════════════════════════════════════"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("                  CPU PERFORMANCE ANALYSIS                         "))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("═══════════════════════════════════════════════════════════════════"))
	b.WriteString("\n\n")

	// Overall CPU Usage section
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render("⚡ CPU OVERVIEW"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n\n")

	// CPU Usage with visual indicator
	usageColor := lipgloss.Color("46") // Green by default
	usageIcon := "●"
	usageStatus := "Normal"
	if metrics.OverallUsage > 80 {
		usageColor = lipgloss.Color("196") // Red for high usage
		usageIcon = "●"
		usageStatus = "High"
	} else if metrics.OverallUsage > 60 {
		usageColor = lipgloss.Color("214") // Orange for medium usage
		usageIcon = "●"
		usageStatus = "Medium"
	}

	usageText := lipgloss.NewStyle().Foreground(usageColor).Bold(true).Render(
		fmt.Sprintf("%s %.1f%% [%s]", usageIcon, metrics.OverallUsage, usageStatus),
	)

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40
	prog.Full = '█'
	prog.Empty = '░'
	usageBar := prog.ViewAs(metrics.OverallUsage / 100.0)

	b.WriteString(fmt.Sprintf("  Overall CPU Usage:  %s\n", usageText))
	b.WriteString(fmt.Sprintf("  %s\n\n", usageBar))

	// Last update and load averages in columns
	lastUpdateText := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(
		fmt.Sprintf("Last Update: %s", metrics.LastUpdate.Format("15:04:05")),
	)
	loadAvgText := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Render(
		fmt.Sprintf("Load Avg: %.2f, %.2f, %.2f", metrics.LoadAverage[0], metrics.LoadAverage[1], metrics.LoadAverage[2]),
	)
	b.WriteString(fmt.Sprintf("  %s    %s\n", lastUpdateText, loadAvgText))
	b.WriteString("\n")

	// CPU State Breakdown
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render("📊 CPU STATE BREAKDOWN"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n\n")

	states := []struct {
		name  string
		value float64
		icon  string
		desc  string
	}{
		{"User", metrics.StateBreakdown.User, "👤", "User space processes"},
		{"System", metrics.StateBreakdown.System, "⚙️ ", "Kernel space processes"},
		{"Idle", metrics.StateBreakdown.Idle, "💤", "CPU idle time"},
		{"IOWait", metrics.StateBreakdown.IOWait, "⏳", "Waiting for I/O operations"},
		{"IRQ", metrics.StateBreakdown.IRQ, "⚡", "Hardware interrupts"},
		{"SoftIRQ", metrics.StateBreakdown.SoftIRQ, "📡", "Software interrupts"},
		{"Steal", metrics.StateBreakdown.Steal, "🔒", "Stolen by hypervisor"},
		{"Nice", metrics.StateBreakdown.Nice, "✨", "Nice priority processes"},
	}

	for _, state := range states {
		stateColor := lipgloss.Color("46")
		if state.value > 20 {
			stateColor = lipgloss.Color("214")
		}
		if state.value > 40 {
			stateColor = lipgloss.Color("196")
		}

		stateProg := progress.New(progress.WithDefaultGradient())
		stateProg.Width = 25
		stateProg.Full = '█'
		stateProg.Empty = '░'
		bar := stateProg.ViewAs(state.value / 100.0)

		nameStyle := lipgloss.NewStyle().Width(10).Foreground(lipgloss.Color("255"))
		valueStyle := lipgloss.NewStyle().Width(8).Foreground(stateColor).Bold(true)

		b.WriteString(fmt.Sprintf("  %s %s %s  %s\n",
			state.icon,
			nameStyle.Render(state.name),
			valueStyle.Render(fmt.Sprintf("%.1f%%", state.value)),
			bar,
		))
	}
	b.WriteString("\n")

	// Per-Core Performance
	if len(metrics.Cores) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render("🔥 PER-CORE PERFORMANCE"))
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", 70))
		b.WriteString("\n\n")

		// Display cores in rows of 2
		for i := 0; i < len(metrics.Cores); i += 2 {
			// First core
			core1 := metrics.Cores[i]
			coreColor1 := lipgloss.Color("46")
			if core1.Usage > 80 {
				coreColor1 = lipgloss.Color("196")
			} else if core1.Usage > 60 {
				coreColor1 = lipgloss.Color("214")
			}

			coreProg1 := progress.New(progress.WithDefaultGradient())
			coreProg1.Width = 15
			coreProg1.Full = '█'
			coreProg1.Empty = '░'
			bar1 := coreProg1.ViewAs(core1.Usage / 100.0)

			coreLabel1 := lipgloss.NewStyle().Width(8).Render(fmt.Sprintf("Core %2d", core1.CoreID))
			coreValue1 := lipgloss.NewStyle().Width(7).Foreground(coreColor1).Bold(true).Render(fmt.Sprintf("%5.1f%%", core1.Usage))

			line := fmt.Sprintf("  %s %s %s", coreLabel1, coreValue1, bar1)

			// Second core (if exists)
			if i+1 < len(metrics.Cores) {
				core2 := metrics.Cores[i+1]
				coreColor2 := lipgloss.Color("46")
				if core2.Usage > 80 {
					coreColor2 = lipgloss.Color("196")
				} else if core2.Usage > 60 {
					coreColor2 = lipgloss.Color("214")
				}

				coreProg2 := progress.New(progress.WithDefaultGradient())
				coreProg2.Width = 15
				coreProg2.Full = '█'
				coreProg2.Empty = '░'
				bar2 := coreProg2.ViewAs(core2.Usage / 100.0)

				coreLabel2 := lipgloss.NewStyle().Width(8).Render(fmt.Sprintf("Core %2d", core2.CoreID))
				coreValue2 := lipgloss.NewStyle().Width(7).Foreground(coreColor2).Bold(true).Render(fmt.Sprintf("%5.1f%%", core2.Usage))

				line += fmt.Sprintf("    %s %s %s", coreLabel2, coreValue2, bar2)
			}

			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	// System Activity Metrics
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Render("📈 SYSTEM ACTIVITY METRICS"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 70))
	b.WriteString("\n\n")

	// Display in two columns
	leftCol := []struct {
		label string
		value string
		icon  string
	}{
		{"Context Switches", formatNumber(metrics.ContextSwitches), "🔄"},
		{"Interrupts", formatNumber(metrics.Interrupts), "⚡"},
		{"Processes", fmt.Sprintf("%d", metrics.ProcessCount), "📋"},
		{"Threads", fmt.Sprintf("%d", metrics.ThreadCount), "🧵"},
	}

	for _, item := range leftCol {
		labelStyle := lipgloss.NewStyle().Width(20).Foreground(lipgloss.Color("255"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
		b.WriteString(fmt.Sprintf("  %s %s %s\n",
			item.icon,
			labelStyle.Render(item.label+":"),
			valueStyle.Render(item.value),
		))
	}

	b.WriteString("\n")

	return b.String()
}

func RenderCPUWithData(metrics *model.CPUMetrics, loading bool) string {
	if loading {
		return vars.CardStyle.Render("⏳ Loading CPU Performance Metrics...")
	}

	if metrics == nil {
		return vars.CardStyle.Render("Press 'r' to load CPU metrics")
	}

	return RenderCPUContent(metrics)
}

func createProgressBar(value float64, color lipgloss.Color) string {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 15
	prog.Full = '█'
	prog.Empty = '░'

	return prog.ViewAs(value / 100.0)
}
